package usecase

import (
	"context"
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
	jwtSecret       string
	accessExpMinute int
	refreshExpDay   int
}

func NewAuthUsecase(ur domain.UserRepository, ar domain.AuthRepository) domain.AuthUsecase {
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
		jwtSecret:       secret,
		accessExpMinute: accessMin,
		refreshExpDay:   refreshDay,
	}
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

	newUser := &domain.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	err = u.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

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

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", "", domain.NewError(domain.ErrBadParamInput, "Email atau password salah")
	}

	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Minute * time.Duration(u.accessExpMinute)).Unix(),
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
