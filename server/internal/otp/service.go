package otp

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrProviderUnavailable = errors.New("otp: sms provider unavailable")

type Service struct {
	Rdb        *redis.Client
	Provider   SMSProvider
	TTL        time.Duration
	HMACSecret []byte
}

func NewService(rdb *redis.Client, p SMSProvider, ttl time.Duration, secret []byte) *Service {
	return &Service{Rdb: rdb, Provider: p, TTL: ttl, HMACSecret: secret}
}

func (s *Service) SendCode(ctx context.Context, phone string) error {
	if s.Provider == nil {
		return ErrProviderUnavailable
	}

	code := randomCode6()
	mac := hmac.New(sha256.New, s.HMACSecret)
	mac.Write([]byte(code))
	sum := mac.Sum(nil)
	key := "otp:" + phone

	if err := s.Rdb.Set(ctx, key, hex.EncodeToString(sum), s.TTL).Err(); err != nil {
		return err
	}

	minutes := int(s.TTL / time.Minute)
	if minutes == 0 {
		minutes = 1
	}

	text := fmt.Sprintf("Kent verification code: %s. Valid for %d min.", code, minutes)
	if err := s.Provider.SendSMS(ctx, phone, text); err != nil {
		_ = s.Rdb.Del(ctx, key).Err()
		if errors.Is(err, ErrProviderNotConfigured) {
			return ErrProviderUnavailable
		}
		return err
	}

	return nil
}

func (s *Service) Verify(ctx context.Context, phone, code string) (bool, error) {
	key := "otp:" + phone
	stored, err := s.Rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, s.HMACSecret)
	mac.Write([]byte(code))
	want := hex.EncodeToString(mac.Sum(nil))
	ok := subtle.ConstantTimeCompare([]byte(stored), []byte(want)) == 1
	if ok {
		_ = s.Rdb.Del(ctx, key).Err()
	}
	return ok, nil
}

func randomCode6() string {
	var n [4]byte
	_, _ = rand.Read(n[:])
	v := int(n[0])<<24 | int(n[1])<<16 | int(n[2])<<8 | int(n[3])
	if v < 0 {
		v = -v
	}
	return fmt.Sprintf("%06d", v%1000000)
}
