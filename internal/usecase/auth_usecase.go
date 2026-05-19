package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
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
	emailSender     domain.EmailSender // <-- TAMBAHAN: Inject EmailSender
	jwtSecret       string
	accessExpMinute int
	refreshExpDay   int
}

// Update constructor untuk menerima domain.EmailSender
func NewAuthUsecase(ur domain.UserRepository, ar domain.AuthRepository, es domain.EmailSender) domain.AuthUsecase {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "omnilibrary-super-secret-key"
	}

	accessMin, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY_MINUTE"))
	if err != nil || accessMin == 0 {
		accessMin = 15
	}

	refreshDay, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_DAY"))
	if err != nil || refreshDay == 0 {
		refreshDay = 7
	}

	return &authUsecase{
		userRepo:        ur,
		authRepo:        ar,
		emailSender:     es, // <-- Inisialisasi EmailSender
		jwtSecret:       secret,
		accessExpMinute: accessMin,
		refreshExpDay:   refreshDay,
	}
}

// Helper function untuk generate token acak yang aman
func generateVerificationToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (u *authUsecase) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
	existingUser, _ := u.userRepo.FindByEmail(ctx, input.Email)
	if existingUser != nil {
		return nil, domain.ErrConflict
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	// 1. Generate Token & Expiry
	token := generateVerificationToken()
	expTime := time.Now().Add(1 * time.Minute) // Berlaku 24 jam

	newUser := &domain.User{
		Name:                  input.Name,
		Email:                 input.Email,
		Password:              string(hashedPassword),
		Role:                  "user",
		IsEmailVerified:       false,    // <-- Set default false
		VerificationToken:     &token,   // <-- Simpan token
		VerificationExpiresAt: &expTime, // <-- Simpan expiry time
	}

	// 2. Simpan ke Database
	err = u.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	// 3. Kirim Email secara Asynchronous (Goroutine)
	// Kita pakai goroutine agar response API tidak perlu menunggu email terkirim
	go func() {
		// Gunakan context.Background() karena context bawaan API (ctx) akan dibatalkan (canceled)
		// saat request selesai, sedangkan goroutine ini mungkin berjalan lebih lama.
		err := u.emailSender.SendVerificationEmail(newUser.Email, token)
		if err != nil {
			// Di produksi, log error ini menggunakan logger (misal: logrus/zap)
			// log.Printf("Gagal mengirim email verifikasi ke %s: %v", newUser.Email, err)
		}
	}()

	return newUser, nil
}

func (u *authUsecase) Login(ctx context.Context, input domain.LoginInput) (string, string, error) {
	user, err := u.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", domain.NewError(domain.ErrNotFound, "User dengan ID tersebut tidak ditemukan")
	}

	// OPSIONAL TAPI SANGAT DISARANKAN:
	// Cek apakah email sudah diverifikasi sebelum mengizinkan login
	// Jika kamu ingin memaksa user verifikasi email dulu, hilangkan comment di bawah ini:
	/*
		if !user.IsEmailVerified {
			return "", "", domain.NewError(domain.ErrUnauthorized, "Silakan verifikasi email Anda terlebih dahulu")
		}
	*/

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", "", domain.NewError(domain.ErrBadParamInput, "Email atau password salah")
	}

	accessClaims := jwt.MapClaims{
		"user_id":           user.ID,
		"role":              user.Role,
		"is_email_verified": user.IsEmailVerified,
		"exp":               time.Now().Add(time.Minute * time.Duration(u.accessExpMinute)).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", "", domain.ErrInternalServerError
	}

	expTime := time.Now().Add(time.Hour * 24 * time.Duration(u.refreshExpDay))
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     expTime.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", "", domain.NewError(domain.ErrUnauthorized, "gagal menerbitkan refresh token")
	}

	rtData := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     signedRefreshToken,
		ExpiresAt: expTime,
	}
	err = u.authRepo.SaveRefreshToken(ctx, rtData)
	if err != nil {
		return "", "", domain.ErrInternalServerError
	}

	return signedAccessToken, signedRefreshToken, nil
}

func (u *authUsecase) Refresh(ctx context.Context, tokenString string) (string, error) {
	// ... (Kode fungsi Refresh tidak ada perubahan) ...
	rt, err := u.authRepo.GetRefreshToken(ctx, tokenString)
	if err != nil {
		return "", err
	}
	if rt == nil {
		return "", domain.NewError(domain.ErrBadParamInput, "Refresh token tidak valid atau sudah kadaluarsa")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode enkripsi tidak valid")
		}
		return []byte(u.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		_ = u.authRepo.DeleteRefreshToken(ctx, tokenString)
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("gagal membaca payload token")
	}
	userID := claims["user_id"].(string)

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return "", domain.NewError(domain.ErrNotFound, "User dengan ID tersebut tidak ditemukan")
	}

	accessClaims := jwt.MapClaims{
		"user_id":           user.ID,
		"role":              user.Role,
		"is_email_verified": user.IsEmailVerified, // <-- Tambahkan informasi verifikasi email di token
		"exp":               time.Now().Add(time.Minute * time.Duration(u.accessExpMinute)).Unix(),
	}
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := newAccessToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", errors.New("gagal menerbitkan access token baru")
	}

	return signedAccessToken, nil
}

// --- IMPLEMENTASI METHOD BARU ---
func (u *authUsecase) VerifyEmail(ctx context.Context, token string) error {
	// 1. Cari user berdasarkan token
	user, err := u.userRepo.FindByVerificationToken(ctx, token)
	if err != nil {
		// Jika error dari DB, kita kembalikan error bad param agar pesan aman untuk user
		return domain.NewError(domain.ErrBadParamInput, "Token verifikasi tidak valid")
	}
	if user == nil {
		return domain.NewError(domain.ErrBadParamInput, "Token verifikasi tidak valid")
	}

	// 2. Cek apakah sudah diverifikasi
	if user.IsEmailVerified {
		return domain.NewError(domain.ErrBadParamInput, "Email sudah diverifikasi sebelumnya")
	}

	// 3. Cek apakah token expired
	if user.VerificationExpiresAt != nil && time.Now().After(*user.VerificationExpiresAt) {
		return domain.NewError(domain.ErrBadParamInput, "Token verifikasi sudah kadaluarsa")
	}

	// 4. Update data user
	user.IsEmailVerified = true
	user.VerificationToken = nil // Hapus (set null) token agar tidak bisa dipakai 2x
	user.VerificationExpiresAt = nil

	// 5. Simpan perubahan ke database
	err = u.userRepo.Update(ctx, user)
	if err != nil {
		return domain.ErrInternalServerError
	}

	return nil
}

func (u *authUsecase) ResendVerification(ctx context.Context, input domain.ResendVerificationInput) error {
	// 1. Cari user berdasarkan email
	user, err := u.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return domain.NewError(domain.ErrNotFound, "User dengan email tersebut tidak ditemukan")
	}
	if user == nil {
		return domain.NewError(domain.ErrNotFound, "User dengan email tersebut tidak ditemukan")
	}

	// 2. Jika email sudah diverifikasi, tidak perlu kirim ulang
	if user.IsEmailVerified {
		return domain.NewError(domain.ErrBadParamInput, "Email sudah diverifikasi sebelumnya")
	}

	// 3. Generate Token baru dan Expiry baru (misal berlaku 24 jam lagi)
	newToken := generateVerificationToken()
	newExpTime := time.Now().Add(24 * time.Hour)

	user.VerificationToken = &newToken
	user.VerificationExpiresAt = &newExpTime

	// 4. Update data user di database
	err = u.userRepo.Update(ctx, user)
	if err != nil {
		return domain.ErrInternalServerError
	}

	// 5. Kirim email secara asynchronous
	go func() {
		err := u.emailSender.SendVerificationEmail(user.Email, newToken)
		if err != nil {
			// Log error jika diperlukan
		}
	}()

	return nil
}

func (u *authUsecase) ForgotPassword(ctx context.Context, input domain.ForgotPasswordInput) error {
	user, err := u.userRepo.FindByEmail(ctx, input.Email)

	// BEST PRACTICE SECURITY:
	// Jangan beritahu user apakah email terdaftar atau tidak (mencegah enumerasi email oleh hacker).
	// Jika user tidak ditemukan, kita tetap return nil (sukses semu).
	if err != nil || user == nil {
		return nil
	}

	// Generate Token & Expiry (15 Menit)
	token := generateVerificationToken() // Kita bisa pakai ulang fungsi helper ini
	expTime := time.Now().Add(15 * time.Minute)

	user.PasswordResetToken = &token
	user.PasswordResetExpiresAt = &expTime

	err = u.userRepo.Update(ctx, user)
	if err != nil {
		return domain.ErrInternalServerError
	}

	// Kirim Email Asynchronous
	go func() {
		_ = u.emailSender.SendPasswordResetEmail(user.Email, token)
	}()

	return nil
}

func (u *authUsecase) ResetPassword(ctx context.Context, input domain.ResetPasswordInput) error {
	// 1. Validasi konfirmasi password
	if input.NewPassword != input.ConfirmPassword {
		return domain.NewError(domain.ErrBadParamInput, "Konfirmasi password tidak cocok")
	}

	// 2. Cari user berdasarkan token reset
	user, err := u.userRepo.FindByResetToken(ctx, input.Token)
	if err != nil || user == nil {
		return domain.NewError(domain.ErrBadParamInput, "Token reset password tidak valid")
	}

	// 3. Cek apakah token expired
	if user.PasswordResetExpiresAt != nil && time.Now().After(*user.PasswordResetExpiresAt) {
		return domain.NewError(domain.ErrBadParamInput, "Token reset password sudah kadaluarsa")
	}

	// 4. Hash password baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrInternalServerError
	}

	// 5. Update data (Mirip seperti di UpdateProfile, kita timpa field-nya)
	user.Password = string(hashedPassword)
	user.PasswordResetToken = nil // Bersihkan token agar tidak bisa dipakai lagi
	user.PasswordResetExpiresAt = nil

	// 6. Simpan ke database menggunakan repository Update yang sama!
	err = u.userRepo.Update(ctx, user)
	if err != nil {
		return domain.ErrInternalServerError
	}

	return nil
}
