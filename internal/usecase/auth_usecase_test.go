package usecase_test

import (
	"context"
	"testing"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/domain/mocks"
	"github.com/faridlan/omni-library-api/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ==========================================
// TEST REGISTER
// ==========================================

func TestRegister_EmailSudahAda(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockAuthRepo := new(mocks.AuthRepository)
	uc := usecase.NewAuthUsecase(mockUserRepo, mockAuthRepo)

	existingUser := &domain.User{Email: "test@example.com"}

	mockAuthRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	user, err := uc.Register(context.Background(), "Faridlan", "test@example.com", "password123")

	assert.ErrorIs(t, err, domain.ErrConflict)
	assert.Nil(t, user)
	mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRegister_Sukses(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockAuthRepo := new(mocks.AuthRepository)
	uc := usecase.NewAuthUsecase(mockUserRepo, mockAuthRepo)

	// Naskah: Saat dicek, email ini BELUM ADA di database (return nil)
	mockAuthRepo.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, nil)

	// Naskah: Pura-pura berhasil menyimpan ke database
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	// Action!
	user, err := uc.Register(context.Background(), "Faridlan", "new@example.com", "password123")

	// Validasi
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "Faridlan", user.Name)
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, "user", user.Role)

	// Validasi Keamanan: Pastikan password yang disimpan BUKAN "password123" asli (sudah di-hash)
	assert.NotEqual(t, "password123", user.Password)

	mockUserRepo.AssertExpectations(t)
}

// ==========================================
// TEST LOGIN
// ==========================================

func TestLogin_EmailTidakDitemukan(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockAuthRepo := new(mocks.AuthRepository)
	uc := usecase.NewAuthUsecase(mockUserRepo, mockAuthRepo)

	mockAuthRepo.On("GetByEmail", mock.Anything, "salah@example.com").Return(nil, nil)

	token, _, err := uc.Login(context.Background(), "salah@example.com", "password123")

	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Empty(t, token)
}

func TestLogin_PasswordSalah(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockAuthRepo := new(mocks.AuthRepository)
	uc := usecase.NewAuthUsecase(mockUserRepo, mockAuthRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password_asli"), bcrypt.DefaultCost)
	dbUser := &domain.User{
		ID:       "user-123",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	mockAuthRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(dbUser, nil)

	token, _, err := uc.Login(context.Background(), "test@example.com", "password_ngawur")

	assert.ErrorIs(t, err, domain.ErrBadParamInput)
	assert.Empty(t, token)
}

func TestLogin_Sukses(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockAuthRepo := new(mocks.AuthRepository)
	uc := usecase.NewAuthUsecase(mockUserRepo, mockAuthRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("rahasia123"), bcrypt.DefaultCost)
	dbUser := &domain.User{
		ID:       "user-123",
		Email:    "test@example.com",
		Role:     "user",
		Password: string(hashedPassword),
	}

	mockAuthRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(dbUser, nil)

	mockAuthRepo.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	token, refreshToken, err := uc.Login(context.Background(), "test@example.com", "rahasia123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, refreshToken)

	assert.True(t, len(token) > 50)
	assert.True(t, len(refreshToken) > 50)
}
