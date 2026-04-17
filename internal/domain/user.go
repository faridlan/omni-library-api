package domain

import (
	"context"
	"time"
)

// User merepresentasikan entitas pengguna di dalam bisnis logika kita
type User struct {
	ID        string
	Name      string
	Email     string
	Password  string // Harus berupa teks yang sudah di-hash (enkripsi), JANGAN PERNAH simpan plain-text!
	Role      string // Contoh: "user" atau "admin"
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepository adalah kontrak untuk tangan yang berinteraksi dengan tabel users
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
}

// AuthUsecase adalah kontrak untuk otak yang mengurus pendaftaran dan login
type AuthUsecase interface {
	Register(ctx context.Context, name, email, password string) (*User, error)

	// Login akan menerima email dan password, lalu mengembalikan token JWT berupa string
	Login(ctx context.Context, email, password string) (string, error)
}
