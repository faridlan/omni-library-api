package usecase

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	userRepo domain.UserRepository
}

func NewAuthUsecase(ur domain.UserRepository) domain.AuthUsecase {
	return &authUsecase{
		userRepo: ur,
	}
}

// Register untuk mendaftarkan Warga baru
func (u *authUsecase) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	// 1. ATURAN BISNIS: Cek apakah email sudah pernah didaftarkan?
	existingUser, _ := u.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, domain.ErrConflict // Menolak jika email kembar
	}

	// 2. KEAMANAN: Acak Password (Hashing) menggunakan algoritma Bcrypt
	// Cost Default (10) adalah standar keseimbangan antara keamanan dan kecepatan
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal mengenkripsi password")
	}

	// 3. RAKIT DATA BARU
	newUser := &domain.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user", // Pendaftar baru selalu menjadi 'user' biasa
	}

	// 4. SIMPAN KE DATABASE LOKAL
	err = u.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// Login untuk mengecek sandi dan menerbitkan Gelang VIP (JWT)
func (u *authUsecase) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Cek apakah emailnya ada di database?
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", domain.ErrNotFound // Email tidak ditemukan
	}

	// 2. Cek apakah Password-nya cocok dengan hash di database?
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Kita kembalikan error "Unauthorized" atau BadParamInput jika password salah
		return "", domain.ErrBadParamInput
	}

	// 3. PEMBUATAN GELANG VIP (JWT)
	// Kita ambil Kunci Rahasia dari file .env (Atau pakai default jika belum ada)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "omnilibrary-super-secret-key"
	}

	// Tentukan isi informasi (Claims) yang akan diselipkan ke dalam Token
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Token berlaku 72 jam (3 hari)
	}

	// Buat token baru menggunakan algoritma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tanda tangani (Sign) token tersebut menggunakan Kunci Rahasia kita
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.New("gagal menerbitkan token")
	}

	return signedToken, nil
}
