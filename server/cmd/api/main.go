package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/kentapp/kent/server/internal/config"
	"github.com/kentapp/kent/server/internal/otp"
)

func main() {
	cfg := config.FromEnv()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}

	provider := otp.NewInfobip(cfg.InfobipBaseURL, cfg.InfobipAPIKey, cfg.InfobipFrom)
	otpSvc := otp.NewService(rdb, provider, cfg.OTPTTL, []byte(cfg.OTPSecret))

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
			Phone string `json:"phone"`
			Code  string `json:"code"`
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

		c.JSON(http.StatusOK, gin.H{"access": "...", "refresh": "..."})
	})

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
