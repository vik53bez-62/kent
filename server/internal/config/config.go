package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrWeakOTPSecret = errors.New("config: OTP_SECRET must be set and at least 32 characters")
	ErrInvalidOTPTTL = errors.New("config: OTP_TTL_SECONDS must be greater than zero")
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

func (c Config) Validate() error {
	secret := strings.TrimSpace(c.OTPSecret)
	if secret == "" || secret == "change_me" || len([]byte(secret)) < 32 {
		return ErrWeakOTPSecret
	}
	if c.OTPTTL <= 0 {
		return ErrInvalidOTPTTL
	}
	return nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
