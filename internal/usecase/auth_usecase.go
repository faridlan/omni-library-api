package usecase

import (
	"context"
	"errors"
	"fmt"
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
		"exp":     time.Now().Add(time.Minute * 1).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", errors.New("gagal menerbitkan access token")
	}

	expTime := time.Now().Add(time.Hour * 24 * 7)
	// 5. PEMBUATAN REFRESH TOKEN (7 Hari)
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     expTime.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", errors.New("gagal menerbitkan refresh token")
	}

	// 6. TODO NANTI: Simpan signedRefreshToken ke Database
	rtData := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     signedRefreshToken,
		ExpiresAt: expTime,
	}
	err = u.userRepo.SaveRefreshToken(ctx, rtData)
	if err != nil {
		return "", "", errors.New("gagal menyimpan sesi login")
	}

	return signedAccessToken, signedRefreshToken, nil
}

func (u *authUsecase) Refresh(ctx context.Context, tokenString string) (string, error) {
	// 1. Cek apakah token ada di Database kita (Belum dicabut/Revoked)
	rt, err := u.userRepo.GetRefreshToken(ctx, tokenString)
	if err != nil {
		return "", err
	}
	if rt == nil {
		return "", domain.ErrBadParamInput // Ditolak: Token tidak ditemukan di DB
	}

	// 2. Parse dan Validasi JWT (Sekaligus mengecek apakah sudah lewat 7 hari)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "omnilibrary-super-secret-key"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode enkripsi tidak valid")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		// Bersih-bersih: Hapus sekalian dari DB jika token sudah kadaluarsa
		_ = u.userRepo.DeleteRefreshToken(ctx, tokenString)
		return "", domain.ErrBadParamInput // Ditolak: Token expired atau invalid
	}

	// 3. Ambil UserID dari dalam token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("gagal membaca payload token")
	}
	userID := claims["user_id"].(string)

	// 4. Ambil data User TERBARU dari database (Penting untuk mengecek Role terkini!)
	user, err := u.userRepo.GetByID(ctx, userID) // Pastikan fungsi GetByID ada di UserRepository kamu
	if err != nil || user == nil {
		return "", domain.ErrNotFound
	}

	// 5. Cetak Access Token BARU (15 Menit)
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role, // Menggunakan role terbaru dari DB!
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	}
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := newAccessToken.SignedString([]byte(secret))
	if err != nil {
		return "", errors.New("gagal menerbitkan access token baru")
	}

	// ==========================================
	// 💡 RUANG KOSONG UNTUK MASA DEPAN (ROTATION)
	// ==========================================
	// Nanti, saat kamu ingin upgrade ke strategi Rotation, buka komen ini:
	/*
		_ = u.authRepo.DeleteRefreshToken(ctx, tokenString)
		newRefreshToken := ... (cetak & simpan ke DB)
		return signedAccessToken, newRefreshToken, nil
	*/

	// Untuk MVP ini (Fixed Strategy), kita cukup kembalikan Access Token baru
	return signedAccessToken, nil
}
