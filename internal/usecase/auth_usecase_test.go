package usecase_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/domain/mocks"
	"github.com/faridlan/omni-library-api/internal/usecase"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Helper untuk setup Usecase dan Mocks (Diperbarui agar mereturn mockEmailSender)
func setupAuthUsecase() (*mocks.UserRepository, *mocks.AuthRepository, *mocks.EmailSender, domain.AuthUsecase) {
	mockUserRepo := new(mocks.UserRepository)
	mockAuthRepo := new(mocks.AuthRepository)
	mockEmail := new(mocks.EmailSender)

	// Set env variables khusus untuk testing
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("ACCESS_TOKEN_EXPIRY_MINUTE", "15")
	os.Setenv("REFRESH_TOKEN_EXPIRY_DAY", "7")

	authUsecase := usecase.NewAuthUsecase(mockUserRepo, mockAuthRepo, mockEmail)
	return mockUserRepo, mockAuthRepo, mockEmail, authUsecase
}

// Helper untuk membuat token JWT valid saat testing Refresh
func generateValidRefreshToken(userID string, secret string) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte(secret))
	return signedToken
}

// ==========================================
// TEST REGISTER
// ==========================================
func TestRegister(t *testing.T) {
	mockUserRepo, _, mockEmailSender, authUsecase := setupAuthUsecase()

	t.Run("Success", func(t *testing.T) {
		input := domain.RegisterInput{
			Name:     "Faridlan",
			Email:    "faridlan@example.com",
			Password: "password123",
		}

		// Mocking: Email belum terdaftar (return nil)
		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(nil, nil).Once()

		// Mocking: Berhasil menyimpan user ke DB
		mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.Email == input.Email && u.Name == input.Name && u.Role == "user"
		})).Return(nil).Once()

		// Mocking: Kirim Email Verifikasi
		mockEmailSender.On("SendVerificationEmail", input.Email, mock.AnythingOfType("string")).Return(nil).Once()

		user, err := authUsecase.Register(context.Background(), input)

		// Jeda agar goroutine SendVerificationEmail sempat dieksekusi
		time.Sleep(50 * time.Millisecond)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "faridlan@example.com", user.Email)

		mockUserRepo.AssertExpectations(t)
		mockEmailSender.AssertExpectations(t)
	})

	t.Run("Failed - Email Already Exists (Conflict)", func(t *testing.T) {
		input := domain.RegisterInput{
			Name:     "Faridlan",
			Email:    "faridlan@example.com",
			Password: "password123",
		}

		existingUser := &domain.User{ID: "123", Email: input.Email}

		// Mocking: Email ditemukan di DB
		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(existingUser, nil).Once()

		user, err := authUsecase.Register(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, domain.ErrConflict, err)
		mockUserRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST LOGIN
// ==========================================
func TestLogin(t *testing.T) {
	mockUserRepo, mockAuthRepo, _, authUsecase := setupAuthUsecase()

	// Setup Hash Password Asli
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	mockUser := &domain.User{
		ID:       "user-123",
		Email:    "faridlan@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}

	t.Run("Success", func(t *testing.T) {
		input := domain.LoginInput{
			Email:    "faridlan@example.com",
			Password: "password123",
		}

		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(mockUser, nil).Once()
		mockAuthRepo.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()

		accessToken, refreshToken, err := authUsecase.Login(context.Background(), input)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})

	t.Run("Failed - User Not Found", func(t *testing.T) {
		input := domain.LoginInput{
			Email:    "unknown@example.com",
			Password: "password123",
		}

		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(nil, nil).Once()

		accessToken, refreshToken, err := authUsecase.Login(context.Background(), input)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Contains(t, err.Error(), "User dengan ID tersebut tidak ditemukan")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failed - Wrong Password", func(t *testing.T) {
		input := domain.LoginInput{
			Email:    "faridlan@example.com",
			Password: "wrongpassword",
		}

		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(mockUser, nil).Once()

		accessToken, refreshToken, err := authUsecase.Login(context.Background(), input)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Contains(t, err.Error(), "Email atau password salah")
		mockUserRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST REFRESH
// ==========================================
func TestRefresh(t *testing.T) {
	mockUserRepo, mockAuthRepo, _, authUsecase := setupAuthUsecase()

	validToken := generateValidRefreshToken("user-123", "test-secret-key")
	mockUser := &domain.User{
		ID:   "user-123",
		Role: "user",
	}

	t.Run("Success", func(t *testing.T) {
		mockRtData := &domain.RefreshToken{Token: validToken, UserID: "user-123"}

		mockAuthRepo.On("GetRefreshToken", mock.Anything, validToken).Return(mockRtData, nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "user-123").Return(mockUser, nil).Once()

		newAccessToken, err := authUsecase.Refresh(context.Background(), validToken)

		assert.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		mockAuthRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failed - Refresh Token Not in DB", func(t *testing.T) {
		mockAuthRepo.On("GetRefreshToken", mock.Anything, validToken).Return(nil, nil).Once()

		newAccessToken, err := authUsecase.Refresh(context.Background(), validToken)

		assert.Error(t, err)
		assert.Empty(t, newAccessToken)
		assert.Contains(t, err.Error(), "Refresh token tidak valid atau sudah kadaluarsa")
		mockAuthRepo.AssertExpectations(t)
	})

	t.Run("Failed - Invalid JWT Signature", func(t *testing.T) {
		invalidToken := generateValidRefreshToken("user-123", "wrong-secret-key")
		mockRtData := &domain.RefreshToken{Token: invalidToken, UserID: "user-123"}

		mockAuthRepo.On("GetRefreshToken", mock.Anything, invalidToken).Return(mockRtData, nil).Once()

		// JWT Parse akan gagal dan memicu penghapusan token
		mockAuthRepo.On("DeleteRefreshToken", mock.Anything, invalidToken).Return(nil).Once()

		newAccessToken, err := authUsecase.Refresh(context.Background(), invalidToken)

		assert.Error(t, err)
		assert.Empty(t, newAccessToken)
		mockAuthRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST VERIFY EMAIL
// ==========================================
func TestVerifyEmail(t *testing.T) {
	mockUserRepo, _, _, authUsecase := setupAuthUsecase()

	t.Run("Success", func(t *testing.T) {
		token := "valid-token-123"
		expTime := time.Now().Add(1 * time.Hour)
		mockUser := &domain.User{
			ID:                    "user-123",
			IsEmailVerified:       false,
			VerificationToken:     &token,
			VerificationExpiresAt: &expTime,
		}

		mockUserRepo.On("FindByVerificationToken", mock.Anything, token).Return(mockUser, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.IsEmailVerified == true && u.VerificationToken == nil && u.VerificationExpiresAt == nil
		})).Return(nil).Once()

		err := authUsecase.VerifyEmail(context.Background(), token)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failed - Invalid Token", func(t *testing.T) {
		token := "invalid-token"
		mockUserRepo.On("FindByVerificationToken", mock.Anything, token).Return(nil, nil).Once()

		err := authUsecase.VerifyEmail(context.Background(), token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Token verifikasi tidak valid")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failed - Token Expired", func(t *testing.T) {
		token := "expired-token"
		// Set waktu kadaluarsa ke 1 jam yang lalu
		expTime := time.Now().Add(-1 * time.Hour)
		mockUser := &domain.User{
			ID:                    "user-123",
			IsEmailVerified:       false,
			VerificationToken:     &token,
			VerificationExpiresAt: &expTime,
		}

		mockUserRepo.On("FindByVerificationToken", mock.Anything, token).Return(mockUser, nil).Once()

		err := authUsecase.VerifyEmail(context.Background(), token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Token verifikasi sudah kadaluarsa")
		mockUserRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST RESEND VERIFICATION
// ==========================================
func TestResendVerification(t *testing.T) {
	mockUserRepo, _, mockEmailSender, authUsecase := setupAuthUsecase()

	t.Run("Success", func(t *testing.T) {
		input := domain.ResendVerificationInput{Email: "faridlan@example.com"}
		mockUser := &domain.User{
			ID:              "user-123",
			Email:           "faridlan@example.com",
			IsEmailVerified: false,
		}

		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(mockUser, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()
		mockEmailSender.On("SendVerificationEmail", input.Email, mock.AnythingOfType("string")).Return(nil).Once()

		err := authUsecase.ResendVerification(context.Background(), input)

		time.Sleep(50 * time.Millisecond) // jeda goroutine

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
		mockEmailSender.AssertExpectations(t)
	})

	t.Run("Failed - User Not Found", func(t *testing.T) {
		input := domain.ResendVerificationInput{Email: "notfound@example.com"}

		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(nil, nil).Once()

		err := authUsecase.ResendVerification(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "User dengan email tersebut tidak ditemukan")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failed - Already Verified", func(t *testing.T) {
		input := domain.ResendVerificationInput{Email: "verified@example.com"}
		mockUser := &domain.User{
			ID:              "user-123",
			Email:           "verified@example.com",
			IsEmailVerified: true, // Sudah diverifikasi
		}

		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(mockUser, nil).Once()

		err := authUsecase.ResendVerification(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Email sudah diverifikasi sebelumnya")
		mockUserRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST FORGOT PASSWORD
// ==========================================
func TestForgotPassword(t *testing.T) {
	mockUserRepo, _, mockEmailSender, authUsecase := setupAuthUsecase()

	t.Run("Success - User Exists", func(t *testing.T) {
		input := domain.ForgotPasswordInput{Email: "faridlan@example.com"}
		mockUser := &domain.User{ID: "user-123", Email: "faridlan@example.com"}

		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(mockUser, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()
		mockEmailSender.On("SendPasswordResetEmail", input.Email, mock.AnythingOfType("string")).Return(nil).Once()

		err := authUsecase.ForgotPassword(context.Background(), input)

		time.Sleep(50 * time.Millisecond) // jeda goroutine
		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
		mockEmailSender.AssertExpectations(t)
	})

	t.Run("Success - User Not Found (Silent Success)", func(t *testing.T) {
		input := domain.ForgotPasswordInput{Email: "unknown@example.com"}

		// Simulasi email tidak ditemukan (return nil, nil)
		mockUserRepo.On("FindByEmail", mock.Anything, input.Email).Return(nil, nil).Once()

		err := authUsecase.ForgotPassword(context.Background(), input)

		// Seharusnya tidak ada error (sesuai security best practice)
		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST RESET PASSWORD
// ==========================================
func TestResetPassword(t *testing.T) {
	mockUserRepo, _, _, authUsecase := setupAuthUsecase()

	t.Run("Success", func(t *testing.T) {
		token := "reset-token-123"
		expTime := time.Now().Add(15 * time.Minute)
		mockUser := &domain.User{
			ID:                     "user-123",
			PasswordResetToken:     &token,
			PasswordResetExpiresAt: &expTime,
		}

		input := domain.ResetPasswordInput{
			Token:           token,
			NewPassword:     "newpassword123",
			ConfirmPassword: "newpassword123",
		}

		mockUserRepo.On("FindByResetToken", mock.Anything, token).Return(mockUser, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		err := authUsecase.ResetPassword(context.Background(), input)

		assert.NoError(t, err)
		assert.Nil(t, mockUser.PasswordResetToken) // Token harus jadi nil setelah reset
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failed - Password Mismatch", func(t *testing.T) {
		input := domain.ResetPasswordInput{
			Token:           "token",
			NewPassword:     "pass1",
			ConfirmPassword: "pass2",
		}

		err := authUsecase.ResetPassword(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Konfirmasi password tidak cocok")
	})

	t.Run("Failed - Token Expired", func(t *testing.T) {
		token := "expired-token"
		expTime := time.Now().Add(-1 * time.Minute)
		mockUser := &domain.User{
			PasswordResetToken:     &token,
			PasswordResetExpiresAt: &expTime,
		}

		input := domain.ResetPasswordInput{
			Token:           token,
			NewPassword:     "newpass123",
			ConfirmPassword: "newpass123",
		}

		mockUserRepo.On("FindByResetToken", mock.Anything, token).Return(mockUser, nil).Once()

		err := authUsecase.ResetPassword(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Token reset password sudah kadaluarsa")
		mockUserRepo.AssertExpectations(t)
	})
}
