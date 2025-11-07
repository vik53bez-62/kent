package main

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "github.com/redis/go-redis/v9"

  "github.com/kentapp/kent/server/internal/config"
  "github.com/kentapp/kent/server/internal/otp"
)

func main() {
  cfg := config.FromEnv()
  rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
  provider := otp.NewInfobip(cfg.InfobipBaseURL, cfg.InfobipAPIKey, cfg.InfobipFrom)
  otpSvc := otp.NewService(rdb, provider, cfg.OTPTTL, []byte(cfg.OTPSecret))

  r := gin.Default()
  r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status":"ok"}) })
  r.POST("/v1/auth/otp/request", func(c *gin.Context) {
    var req struct {
      Phone string `json:"phone"`
    }
    if c.BindJSON(&req) != nil || req.Phone == "" {
      c.JSON(http.StatusBadRequest, gin.H{"error":"bad_request"})
      return
    }
    if err := otpSvc.SendCode(c, req.Phone); err != nil {
      c.JSON(http.StatusBadGateway, gin.H{"error":"otp_failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"ok": true})
  })
  r.POST("/v1/auth/otp/verify", func(c *gin.Context) {
    var req struct {
      Phone string `json:"phone"`
      Code  string `json:"code"`
    }
    if c.BindJSON(&req) != nil || req.Phone == "" || req.Code == "" {
      c.JSON(http.StatusBadRequest, gin.H{"error":"bad_request"})
      return
    }
    ok, err := otpSvc.Verify(c, req.Phone, req.Code)
    if err != nil || !ok {
      c.JSON(http.StatusUnauthorized, gin.H{"error":"invalid_code"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"access":"...", "refresh":"..."})
  })
  _ = r.Run(":" + cfg.Port)
}
