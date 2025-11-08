package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/kentapp/kent/server/internal/auth"
	"github.com/kentapp/kent/server/internal/config"
	"github.com/kentapp/kent/server/internal/db"
	"github.com/kentapp/kent/server/internal/otp"
	"github.com/kentapp/kent/server/internal/users"
)

const claimsKey = "claims"

func main() {
	cfg := config.FromEnv()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}

	provider := otp.NewInfobip(cfg.InfobipBaseURL, cfg.InfobipAPIKey, cfg.InfobipFrom)
	otpSvc := otp.NewService(rdb, provider, cfg.OTPTTL, []byte(cfg.OTPSecret))

	userRepo := users.NewRepository(pool)
	authSvc := auth.NewService(pool, []byte(cfg.AccessSecret), []byte(cfg.RefreshSecret), cfg.AccessTTL, cfg.RefreshTTL)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/v1/auth/otp/request", func(c *gin.Context) {
		var req struct {
			Phone string `json:"phone"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		phone := strings.TrimSpace(req.Phone)
		if phone == "" || len(phone) > 32 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		if err := otpSvc.SendCode(c.Request.Context(), phone); err != nil {
			if errors.Is(err, otp.ErrProviderUnavailable) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "otp_unavailable"})
				return
			}
			log.Printf("otp send failed: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "otp_failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.POST("/v1/auth/otp/verify", func(c *gin.Context) {
		var req struct {
			Phone       string `json:"phone"`
			Code        string `json:"code"`
			DeviceLabel string `json:"deviceLabel"`
			PushToken   string `json:"pushToken"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		phone := strings.TrimSpace(req.Phone)
		code := strings.TrimSpace(req.Code)
		if phone == "" || code == "" || len(phone) > 32 || len(code) > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		ok, err := otpSvc.Verify(c.Request.Context(), phone, code)
		if err != nil {
			log.Printf("otp verify failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "otp_verify_failed"})
			return
		}
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_code"})
			return
		}

		label := optionalString(req.DeviceLabel)
		push := optionalString(req.PushToken)

		user, device, err := userRepo.UpsertUserAndDevice(c.Request.Context(), phone, label, push)
		if err != nil {
			log.Printf("user upsert failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth_failed"})
			return
		}

		tokens, err := authSvc.IssueTokens(
			c.Request.Context(),
			user.ID,
			device.ID,
			c.Request.UserAgent(),
			c.ClientIP(),
		)
		if err != nil {
			log.Printf("issue tokens failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth_failed"})
			return
		}

		resp := gin.H{
			"user": gin.H{
				"id":          user.ID,
				"phone":       user.Phone,
				"displayName": user.DisplayName,
			},
			"access": gin.H{
				"token":     tokens.AccessToken,
				"expiresAt": tokens.AccessExpiresAt,
			},
			"refresh": gin.H{
				"token":     tokens.RefreshToken,
				"expiresAt": tokens.RefreshExpiresAt,
				"sessionId": tokens.SessionID,
			},
			"device": gin.H{
				"id":    device.ID,
				"label": device.Label,
			},
		}

		c.JSON(http.StatusOK, resp)
	})

	r.POST("/v1/auth/token/refresh", func(c *gin.Context) {
		var req struct {
			Refresh string `json:"refresh"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Refresh) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		tokens, err := authSvc.RefreshTokens(c.Request.Context(), strings.TrimSpace(req.Refresh), c.Request.UserAgent(), c.ClientIP())
		if err != nil {
			if errors.Is(err, auth.ErrInvalidRefreshToken) || errors.Is(err, auth.ErrSessionNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
				return
			}
			log.Printf("refresh failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh_failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access": gin.H{
				"token":     tokens.AccessToken,
				"expiresAt": tokens.AccessExpiresAt,
			},
			"refresh": gin.H{
				"token":     tokens.RefreshToken,
				"expiresAt": tokens.RefreshExpiresAt,
				"sessionId": tokens.SessionID,
			},
		})
	})

	authGroup := r.Group("/v1")
	authGroup.Use(authMiddleware(authSvc))

	authGroup.POST("/auth/logout", func(c *gin.Context) {
		var req struct {
			Refresh string `json:"refresh"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Refresh) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		claims := c.MustGet(claimsKey).(*auth.Claims)
		sessionID, err := parseSessionID(req.Refresh)
		if err != nil || claims.SessionID != sessionID.String() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
			return
		}

		if err := authSvc.RevokeSession(c.Request.Context(), strings.TrimSpace(req.Refresh)); err != nil {
			if errors.Is(err, auth.ErrSessionNotFound) {
				c.JSON(http.StatusGone, gin.H{"error": "session_expired"})
				return
			}
			log.Printf("logout failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "logout_failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	authGroup.GET("/users/me", func(c *gin.Context) {
		claims := c.MustGet(claimsKey).(*auth.Claims)
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			log.Printf("get user failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_lookup_failed"})
			return
		}
		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          user.ID,
			"phone":       user.Phone,
			"displayName": user.DisplayName,
			"createdAt":   user.CreatedAt,
		})
	})

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func authMiddleware(authSvc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		token := strings.TrimSpace(header[7:])
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := authSvc.ParseAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(claimsKey, claims)
		c.Next()
	}
}

func optionalString(v string) *string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func parseSessionID(refresh string) (uuid.UUID, error) {
	parts := strings.Split(refresh, ".")
	if len(parts) != 2 {
		return uuid.Nil, errors.New("invalid")
	}
	return uuid.Parse(parts[0])
}
