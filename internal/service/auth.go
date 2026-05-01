package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost         = 12
	defaultBaseCurrency = "KZT"
)

type RegisterInput struct {
	Email        string
	Password     string
	BaseCurrency string
}

type LoginInput struct {
	Email    string
	Password string
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type Auth interface {
	Register(ctx context.Context, in RegisterInput) (*TokenResponse, *domain.User, error)
	Login(ctx context.Context, in LoginInput) (*TokenResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error)
	Logout(ctx context.Context, accessToken string, refreshToken *string) error
	Me(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	TokenManager() TokenManager
}

type AuthService struct {
	users        repository.User
	tokenManager TokenManager
	onRegister   []func(ctx context.Context, user *domain.User) error
}

func NewAuthService(users repository.User, tm TokenManager) *AuthService {
	return &AuthService{
		users:        users,
		tokenManager: tm,
	}
}

// RegisterHook adds a callback executed after a user is created.
// Used so that other services (e.g. categories) can seed defaults without
// the auth service depending on them directly.
func (s *AuthService) RegisterHook(hook func(ctx context.Context, user *domain.User) error) {
	s.onRegister = append(s.onRegister, hook)
}

func (s *AuthService) TokenManager() TokenManager {
	return s.tokenManager
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*TokenResponse, *domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.Password == "" {
		return nil, nil, domain.ErrInvalidInput
	}
	currency := strings.ToUpper(strings.TrimSpace(in.BaseCurrency))
	if currency == "" {
		currency = defaultBaseCurrency
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
		BaseCurrency: currency,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	for _, hook := range s.onRegister {
		if err := hook(ctx, user); err != nil {
			return nil, nil, fmt.Errorf("post-register hook: %w", err)
		}
	}

	pair, err := s.tokenManager.GeneratePair(user.ID)
	if err != nil {
		return nil, nil, err
	}
	return tokenPairToResponse(pair), user, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*TokenResponse, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	pair, err := s.tokenManager.GeneratePair(user.ID)
	if err != nil {
		return nil, err
	}
	return tokenPairToResponse(pair), nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	claims, err := s.tokenManager.ParseRefresh(refreshToken)
	if err != nil {
		return nil, err
	}
	revoked, err := s.tokenManager.IsRefreshRevoked(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("check revocation: %w", err)
	}
	if revoked {
		return nil, domain.ErrInvalidToken
	}

	if err := s.tokenManager.RevokeRefresh(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
		return nil, fmt.Errorf("revoke old refresh: %w", err)
	}

	pair, err := s.tokenManager.GeneratePair(claims.UserID)
	if err != nil {
		return nil, err
	}
	return tokenPairToResponse(pair), nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken string, refreshToken *string) error {
	claims, err := s.tokenManager.ParseAccess(accessToken)
	if err != nil {
		return domain.ErrInvalidToken
	}
	if claims.ID != "" {
		if err := s.tokenManager.BlacklistAccess(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
			return fmt.Errorf("blacklist access token: %w", err)
		}
	}
	if refreshToken != nil && *refreshToken != "" {
		rc, err := s.tokenManager.ParseRefresh(*refreshToken)
		if err == nil && rc.ID != "" {
			_ = s.tokenManager.RevokeRefresh(ctx, rc.ID, rc.ExpiresAt.Time)
		}
	}
	return nil
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, userID)
}

func tokenPairToResponse(p *TokenPair) *TokenResponse {
	return &TokenResponse{
		AccessToken:      p.AccessToken,
		RefreshToken:     p.RefreshToken,
		AccessExpiresAt:  p.AccessExpiresAt,
		RefreshExpiresAt: p.RefreshExpiresAt,
	}
}
