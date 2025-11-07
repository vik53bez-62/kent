package config

import (
  "os"
  "strconv"
  "time"
)

type Config struct {
  InfobipBaseURL string
  InfobipAPIKey  string
  InfobipFrom    string
  OTPSecret      string
  OTPTTL         time.Duration
  RedisAddr      string
  Port           string
}

func FromEnv() Config {
  ttl, _ := strconv.Atoi(getenv("OTP_TTL_SECONDS", "300"))
  return Config{
    InfobipBaseURL: getenv("INFOBIP_BASE_URL", ""),
    InfobipAPIKey:  getenv("INFOBIP_API_KEY", ""),
    InfobipFrom:    getenv("INFOBIP_FROM", "KENT"),
    OTPSecret:      getenv("OTP_SECRET", "change_me"),
    OTPTTL:         time.Duration(ttl) * time.Second,
    RedisAddr:      getenv("REDIS_ADDR", "localhost:6379"),
    Port:           getenv("PORT", "8080"),
  }
}

func getenv(k, d string) string {
  if v := os.Getenv(k); v != "" {
    return v
  }
  return d
}
