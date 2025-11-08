package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidRefreshToken = errors.New("auth: invalid refresh token")
	ErrSessionNotFound     = errors.New("auth: session not found")
)

type Service struct {
	pool          *pgxpool.Pool
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

type Tokens struct {
	AccessToken      string    `json:"access"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshToken     string    `json:"refresh"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
	UserID           uuid.UUID `json:"userId"`
	DeviceID         uuid.UUID `json:"deviceId"`
	SessionID        uuid.UUID `json:"sessionId"`
}

type Claims struct {
	UserID    string `json:"uid"`
	DeviceID  string `json:"did"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

func NewService(pool *pgxpool.Pool, accessSecret, refreshSecret []byte, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		pool:          pool,
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (s *Service) IssueTokens(ctx context.Context, userID, deviceID uuid.UUID, userAgent, ip string) (*Tokens, error) {
	sessionID := uuid.New()
	refreshToken, refreshHash, err := s.generateRefreshToken(sessionID)
	if err != nil {
		return nil, err
	}

	refreshExpiresAt := time.Now().Add(s.refreshTTL)
	_, err = s.pool.Exec(ctx, `INSERT INTO sessions (id, user_id, device_id, refresh_token_sha256, expires_at, user_agent, ip)
      VALUES ($1, $2, $3, $4, $5, $6, $7)
      ON CONFLICT (id) DO UPDATE SET refresh_token_sha256=EXCLUDED.refresh_token_sha256, expires_at=EXCLUDED.expires_at, updated_at=now(), last_used_at=now(), user_agent=EXCLUDED.user_agent, ip=EXCLUDED.ip`,
		sessionID, userID, deviceID, refreshHash, refreshExpiresAt, userAgent, ip,
	)
	if err != nil {
		return nil, err
	}

	accessToken, accessExp, err := s.signAccessToken(userID, deviceID, sessionID)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExp,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
		UserID:           userID,
		DeviceID:         deviceID,
		SessionID:        sessionID,
	}, nil
}

func (s *Service) RefreshTokens(ctx context.Context, token string, userAgent, ip string) (*Tokens, error) {
	sessionID, hash, err := s.parseAndHashRefresh(token)
	if err != nil {
		return nil, err
	}

	var userID, deviceID uuid.UUID
	var storedHash []byte
	var expiresAt time.Time
	err = s.pool.QueryRow(ctx, `SELECT user_id, device_id, refresh_token_sha256, expires_at FROM sessions WHERE id=$1`, sessionID).Scan(&userID, &deviceID, &storedHash, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if time.Now().After(expiresAt) {
		_, _ = s.pool.Exec(ctx, `DELETE FROM sessions WHERE id=$1`, sessionID)
		return nil, ErrInvalidRefreshToken
	}

	if !hmac.Equal(storedHash, hash) {
		return nil, ErrInvalidRefreshToken
	}

	refreshToken, refreshHash, err := s.generateRefreshToken(sessionID)
	if err != nil {
		return nil, err
	}
	refreshExpiresAt := time.Now().Add(s.refreshTTL)

	_, err = s.pool.Exec(ctx, `UPDATE sessions SET refresh_token_sha256=$1, expires_at=$2, updated_at=now(), last_used_at=now(), user_agent=$3, ip=$4 WHERE id=$5`, refreshHash, refreshExpiresAt, userAgent, ip, sessionID)
	if err != nil {
		return nil, err
	}

	accessToken, accessExp, err := s.signAccessToken(userID, deviceID, sessionID)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExp,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
		UserID:           userID,
		DeviceID:         deviceID,
		SessionID:        sessionID,
	}, nil
}

func (s *Service) RevokeSession(ctx context.Context, refreshToken string) error {
	sessionID, hash, err := s.parseAndHashRefresh(refreshToken)
	if err != nil {
		return err
	}

	res, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE id=$1 AND refresh_token_sha256=$2`, sessionID, hash)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (s *Service) ParseAccessToken(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidRefreshToken
	}
	return claims, nil
}

func (s *Service) signAccessToken(userID, deviceID, sessionID uuid.UUID) (string, time.Time, error) {
	now := time.Now()
	expires := now.Add(s.accessTTL)
	claims := &Claims{
		UserID:    userID.String(),
		DeviceID:  deviceID.String(),
		SessionID: sessionID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.accessSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expires, nil
}

func (s *Service) generateRefreshToken(sessionID uuid.UUID) (string, []byte, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, err
	}
	raw := sessionID.String() + "." + base64.RawURLEncoding.EncodeToString(buf)
	hash := hashValue(raw, s.refreshSecret)
	return raw, hash, nil
}

func (s *Service) parseAndHashRefresh(token string) (uuid.UUID, []byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return uuid.Nil, nil, ErrInvalidRefreshToken
	}
	sessionID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, nil, ErrInvalidRefreshToken
	}
	hash := hashValue(token, s.refreshSecret)
	return sessionID, hash, nil
}

func hashValue(raw string, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(raw))
	return mac.Sum(nil)
}
