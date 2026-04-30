package dto

import (
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/service"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	BaseCurrency string `json:"base_currency" binding:"omitempty,len=3"`
}

func (r RegisterRequest) ToInput() service.RegisterInput {
	return service.RegisterInput{
		Email:        r.Email,
		Password:     r.Password,
		BaseCurrency: r.BaseCurrency,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (r LoginRequest) ToInput() service.LoginInput {
	return service.LoginInput{
		Email:    r.Email,
		Password: r.Password,
	}
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

func TokenResponseFrom(t *service.TokenResponse) TokenResponse {
	return TokenResponse{
		AccessToken:      t.AccessToken,
		RefreshToken:     t.RefreshToken,
		AccessExpiresAt:  t.AccessExpiresAt,
		RefreshExpiresAt: t.RefreshExpiresAt,
	}
}

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	BaseCurrency string    `json:"base_currency"`
	CreatedAt    time.Time `json:"created_at"`
}

func UserResponseFrom(u *domain.User) UserResponse {
	return UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		BaseCurrency: u.BaseCurrency,
		CreatedAt:    u.CreatedAt,
	}
}
