package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	userRepo        domain.UserRepository
	authRepo        domain.AuthRepository
	jwtSecret       string
	accessExpMinute int
	refreshExpDay   int
}

func NewAuthUsecase(ur domain.UserRepository, ar domain.AuthRepository) domain.AuthUsecase {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "omnilibrary-super-secret-key"
	}

	// 2. Baca Access Expiry
	accessMin, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY_MINUTE"))
	if err != nil || accessMin == 0 {
		accessMin = 15 // Default
	}

	// 3. Baca Refresh Expiry
	refreshDay, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_DAY"))
	if err != nil || refreshDay == 0 {
		refreshDay = 7 // Default
	}

	return &authUsecase{
		userRepo:        ur,
		authRepo:        ar,
		jwtSecret:       secret,
		accessExpMinute: accessMin,
		refreshExpDay:   refreshDay,
	}
}

// Register untuk mendaftarkan Warga baru
func (u *authUsecase) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	// 1. ATURAN BISNIS: Cek apakah email sudah pernah didaftarkan?
	existingUser, _ := u.authRepo.GetByEmail(ctx, email)
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
	user, err := u.authRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", domain.ErrNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", domain.ErrBadParamInput
	}

	// 3. PEMBUATAN ACCESS TOKEN (Gunakan u.accessExpMinute dan u.jwtSecret)
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Minute * time.Duration(u.accessExpMinute)).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", "", errors.New("gagal menerbitkan access token")
	}

	// 4. PEMBUATAN REFRESH TOKEN (Gunakan u.refreshExpDay)
	expTime := time.Now().Add(time.Hour * 24 * time.Duration(u.refreshExpDay))
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     expTime.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", "", errors.New("gagal menerbitkan refresh token")
	}

	// 5. SIMPAN KE DATABASE
	rtData := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     signedRefreshToken,
		ExpiresAt: expTime,
	}
	err = u.authRepo.SaveRefreshToken(ctx, rtData)
	if err != nil {
		return "", "", errors.New("gagal menyimpan sesi login")
	}

	return signedAccessToken, signedRefreshToken, nil
}

func (u *authUsecase) Refresh(ctx context.Context, tokenString string) (string, error) {
	// 1. Cek apakah token ada di Database kita (Belum dicabut/Revoked)
	rt, err := u.authRepo.GetRefreshToken(ctx, tokenString)
	if err != nil {
		return "", err
	}
	if rt == nil {
		return "", domain.ErrBadParamInput // Ditolak: Token tidak ditemukan di DB
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode enkripsi tidak valid")
		}
		return []byte(u.jwtSecret), nil // <-- Jauh lebih rapi
	})

	if err != nil || !token.Valid {
		_ = u.authRepo.DeleteRefreshToken(ctx, tokenString)
		return "", domain.ErrBadParamInput
	}

	// 3. Ambil UserID dari dalam token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("gagal membaca payload token")
	}
	userID := claims["user_id"].(string)

	// 4. Ambil data User TERBARU dari database (Penting untuk mengecek Role terkini!)
	user, err := u.authRepo.GetByID(ctx, userID) // Pastikan fungsi GetByID ada di UserRepository kamu
	if err != nil || user == nil {
		return "", domain.ErrNotFound
	}

	// 5. Cetak Access Token BARU (15 Menit)
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Minute * time.Duration(u.accessExpMinute)).Unix(),
	}
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := newAccessToken.SignedString([]byte(u.jwtSecret))
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
