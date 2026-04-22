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
func (u *authUsecase) Login(ctx context.Context, email, password string) (string, string, error) {
	// 1. Cek User
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", domain.ErrNotFound
	}

	// 2. Cek Password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", domain.ErrBadParamInput
	}

	// 3. Ambil Secret Key
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "omnilibrary-super-secret-key"
	}

	// 4. PEMBUATAN ACCESS TOKEN (15 Menit)
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", errors.New("gagal menerbitkan access token")
	}

	// 5. PEMBUATAN REFRESH TOKEN (7 Hari)
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", errors.New("gagal menerbitkan refresh token")
	}

	// 6. TODO NANTI: Simpan signedRefreshToken ke Database
	// err = u.authRepo.SaveRefreshToken(ctx, user.ID, signedRefreshToken, expTime)
	// ...

	return signedAccessToken, signedRefreshToken, nil
}
