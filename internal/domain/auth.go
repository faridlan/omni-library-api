package domain

import (
	"context"
	"time"
)

// Struct representasi data Refresh Token
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginInput struct {
	Email    string
	Password string
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, rt *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
type AuthUsecase interface {
	Register(ctx context.Context, input RegisterInput) (*User, error)
	Login(ctx context.Context, input LoginInput) (string, string, error)
	Refresh(ctx context.Context, tokenString string) (string, error)
}
