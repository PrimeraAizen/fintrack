package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/diyas/fintrack/config"
	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	RefreshJTI       string
}

type TokenManager interface {
	GeneratePair(userID uuid.UUID) (*TokenPair, error)
	ParseAccess(token string) (*Claims, error)
	ParseRefresh(token string) (*Claims, error)
	RevokeRefresh(ctx context.Context, jti string, expiresAt time.Time) error
	IsRefreshRevoked(ctx context.Context, jti string) (bool, error)
}

type tokenManager struct {
	cfg   *config.JWT
	redis *postgres.RedisClient
}

func NewTokenManager(cfg *config.JWT, rdb *postgres.RedisClient) TokenManager {
	return &tokenManager{cfg: cfg, redis: rdb}
}

func (t *tokenManager) GeneratePair(userID uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(t.cfg.AccessTTL)
	refreshExp := now.Add(t.cfg.RefreshTTL)
	refreshJTI := uuid.NewString()

	accessTok, err := t.sign(Claims{
		UserID:    userID,
		TokenType: tokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExp),
			Subject:   userID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshTok, err := t.sign(Claims{
		UserID:    userID,
		TokenType: tokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			Subject:   userID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:      accessTok,
		RefreshToken:     refreshTok,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
		RefreshJTI:       refreshJTI,
	}, nil
}

func (t *tokenManager) sign(claims Claims) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(t.cfg.Secret))
}

func (t *tokenManager) parse(tokenStr, expectedType string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(t.cfg.Secret), nil
	})
	if err != nil || !tok.Valid {
		return nil, domain.ErrInvalidToken
	}
	if claims.TokenType != expectedType {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}

func (t *tokenManager) ParseAccess(token string) (*Claims, error) {
	return t.parse(token, tokenTypeAccess)
}

func (t *tokenManager) ParseRefresh(token string) (*Claims, error) {
	return t.parse(token, tokenTypeRefresh)
}

func (t *tokenManager) RevokeRefresh(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return t.redis.Client.Set(ctx, refreshBlacklistKey(jti), "1", ttl).Err()
}

func (t *tokenManager) IsRefreshRevoked(ctx context.Context, jti string) (bool, error) {
	val, err := t.redis.Client.Get(ctx, refreshBlacklistKey(jti)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

func refreshBlacklistKey(jti string) string {
	return "refresh_blacklist:" + jti
}
