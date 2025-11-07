package otp

import (
  "context"
  "crypto/hmac"
  "crypto/rand"
  "crypto/sha256"
  "crypto/subtle"
  "encoding/hex"
  "fmt"
  "time"

  "github.com/redis/go-redis/v9"
)

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
  code := randomCode6()
  mac := hmac.New(sha256.New, s.HMACSecret)
  mac.Write([]byte(code))
  sum := mac.Sum(nil)
  key := "otp:" + phone
  if err := s.Rdb.Set(ctx, key, hex.EncodeToString(sum), s.TTL).Err(); err != nil {
    return err
  }
  text := fmt.Sprintf("Kent: %s — код подтверждения. Срок действия %d мин.", code, int(s.TTL.Minutes()))
  return s.Provider.SendSMS(ctx, phone, text)
}

func (s *Service) Verify(ctx context.Context, phone, code string) (bool, error) {
  key := "otp:" + phone
  stored, err := s.Rdb.Get(ctx, key).Result()
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
