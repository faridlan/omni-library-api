package postgres

import (
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/google/uuid"
)

// UserModel merepresentasikan tabel users yang sudah di-update
type UserModel struct {
	ID           string `gorm:"type:uuid;primary_key"`
	Name         string `gorm:"type:varchar(255);not null"`
	Email        string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null"` // <-- PENYESUAIAN NAMA KOLOM SQL
	Role         string `gorm:"type:varchar(50);not null;default:'user'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserModel) TableName() string {
	return "users"
}

// 1. Model GORM untuk pemetaan ke tabel 'refresh_tokens'
type RefreshTokenModel struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `gorm:"type:uuid;not null"`
	Token     string    `gorm:"type:text;unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// Beri tahu GORM nama tabel pastinya
func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

func (m *UserModel) ToDomain() *domain.User {
	return &domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.PasswordHash, // Petakan PasswordHash dari DB ke Password di Domain
		Role:      m.Role,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func FromUserDomain(u *domain.User) *UserModel {
	id := u.ID
	if id == "" {
		id = uuid.NewString()
	}

	return &UserModel{
		ID:           id,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.Password, // Petakan Password dari Domain ke PasswordHash di DB
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
