package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrWeakOTPSecret      = errors.New("config: OTP_SECRET must be set and at least 32 characters")
	ErrInvalidOTPTTL      = errors.New("config: OTP_TTL_SECONDS must be greater than zero")
	ErrMissingDatabaseURL = errors.New("config: DATABASE_URL must be provided")
	ErrWeakAccessSecret   = errors.New("config: ACCESS_SECRET must be at least 32 characters")
	ErrWeakRefreshSecret  = errors.New("config: REFRESH_SECRET must be at least 32 characters")
)

type Config struct {
	InfobipBaseURL string
	InfobipAPIKey  string
	InfobipFrom    string
	OTPSecret      string
	OTPTTL         time.Duration
	RedisAddr      string
	Port           string
	DatabaseURL    string
	AccessSecret   string
	RefreshSecret  string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
}

func FromEnv() Config {
	ttlSeconds, _ := strconv.Atoi(getenv("OTP_TTL_SECONDS", "300"))
	accessTTL := parseDuration(getenv("ACCESS_TOKEN_TTL", "15m"), 15*time.Minute)
	refreshTTL := parseDuration(getenv("REFRESH_TOKEN_TTL", "720h"), 30*24*time.Hour)

	return Config{
		InfobipBaseURL: getenv("INFOBIP_BASE_URL", ""),
		InfobipAPIKey:  getenv("INFOBIP_API_KEY", ""),
		InfobipFrom:    getenv("INFOBIP_FROM", "KENT"),
		OTPSecret:      getenv("OTP_SECRET", "change_me"),
		OTPTTL:         time.Duration(ttlSeconds) * time.Second,
		RedisAddr:      getenv("REDIS_ADDR", "localhost:6379"),
		Port:           getenv("PORT", "8080"),
		DatabaseURL:    getenv("DATABASE_URL", ""),
		AccessSecret:   getenv("ACCESS_SECRET", "change_me_access"),
		RefreshSecret:  getenv("REFRESH_SECRET", "change_me_refresh"),
		AccessTTL:      accessTTL,
		RefreshTTL:     refreshTTL,
	}
}

func (c Config) Validate() error {
	if err := validateSecret(c.OTPSecret, ErrWeakOTPSecret); err != nil {
		return err
	}
	if c.OTPTTL <= 0 {
		return ErrInvalidOTPTTL
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return ErrMissingDatabaseURL
	}
	if err := validateSecret(c.AccessSecret, ErrWeakAccessSecret); err != nil {
		return err
	}
	if err := validateSecret(c.RefreshSecret, ErrWeakRefreshSecret); err != nil {
		return err
	}
	if c.AccessTTL <= 0 {
		return fmt.Errorf("access ttl must be > 0")
	}
	if c.RefreshTTL <= 0 || c.RefreshTTL < c.AccessTTL {
		return fmt.Errorf("refresh ttl must be > access ttl")
	}
	return nil
}

func validateSecret(secret string, weakErr error) error {
	s := strings.TrimSpace(secret)
	if s == "" || len([]byte(s)) < 32 || strings.Contains(s, "change_me") {
		return weakErr
	}
	return nil
}

func parseDuration(raw string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return d
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
