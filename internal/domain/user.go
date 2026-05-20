package domain

import (
	"context"
	"time"
)

type User struct {
	ID                     string
	Name                   string
	Email                  string
	Password               string
	Role                   string
	IsEmailVerified        bool
	VerificationToken      *string
	VerificationExpiresAt  *time.Time
	PasswordResetToken     *string
	PasswordResetExpiresAt *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type UpdatePasswordInput struct {
	ID          string
	OldPassword string
	NewPassword string
}

type UpdateProfileInput struct {
	ID   string
	Name string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, user *User) error
	FindByVerificationToken(ctx context.Context, token string) (*User, error)
	FindByResetToken(ctx context.Context, token string) (*User, error)
}

type UserUsecase interface {
	GetProfile(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, input UpdateProfileInput) (*User, error)
	UpdatePassword(ctx context.Context, input UpdatePasswordInput) error
}
