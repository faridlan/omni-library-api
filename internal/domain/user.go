package domain

import (
	"context"
	"time"

	"github.com/faridlan/omni-library-api/internal/delivery/http/dto"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Jangan pernah return password
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, user *User) error
}

type UserUsecase interface {
	GetProfile(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) (*User, error)
	UpdatePassword(ctx context.Context, userID string, req *dto.UpdatePasswordRequest) error
}
