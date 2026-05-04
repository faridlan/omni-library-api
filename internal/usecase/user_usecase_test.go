package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/domain/mocks"
	"github.com/faridlan/omni-library-api/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Helper untuk setup Usecase dan Mocks
func setupUserUsecase() (*mocks.UserRepository, domain.UserUsecase) {
	mockUserRepo := new(mocks.UserRepository)
	userUsecase := usecase.NewUserUsecase(mockUserRepo)
	return mockUserRepo, userUsecase
}

// ==========================================
// TEST GET PROFILE
// ==========================================
func TestGetProfile(t *testing.T) {
	mockRepo, userUsecase := setupUserUsecase()
	userID := "user-123"

	t.Run("Success", func(t *testing.T) {
		mockUser := &domain.User{
			ID:    userID,
			Name:  "Faridlan",
			Email: "faridlan@example.com",
		}

		mockRepo.On("FindByID", mock.Anything, userID).Return(mockUser, nil).Once()

		user, err := userUsecase.GetProfile(context.Background(), userID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Faridlan", user.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - User Not Found", func(t *testing.T) {
		mockRepo.On("FindByID", mock.Anything, userID).Return(nil, domain.ErrNotFound).Once()

		user, err := userUsecase.GetProfile(context.Background(), userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "User dengan ID tersebut tidak ditemukan")
		assert.ErrorIs(t, err, domain.ErrNotFound) // Memastikan Unwrap() berjalan dengan baik
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - Database Error", func(t *testing.T) {
		dbError := errors.New("database connection lost")
		mockRepo.On("FindByID", mock.Anything, userID).Return(nil, dbError).Once()

		user, err := userUsecase.GetProfile(context.Background(), userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST UPDATE PROFILE
// ==========================================
func TestUpdateProfile(t *testing.T) {
	mockRepo, userUsecase := setupUserUsecase()

	t.Run("Success", func(t *testing.T) {
		input := domain.UpdateProfileInput{
			ID:   "user-123",
			Name: "Faridlan Updated",
		}
		mockUser := &domain.User{ID: input.ID, Name: "Faridlan Lama"}

		mockRepo.On("FindByID", mock.Anything, input.ID).Return(mockUser, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.Name == input.Name // Memastikan Name benar-benar diubah sebelum disave
		})).Return(nil).Once()

		user, err := userUsecase.UpdateProfile(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Faridlan Updated", user.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - User Not Found", func(t *testing.T) {
		input := domain.UpdateProfileInput{ID: "user-123", Name: "Faridlan Updated"}

		mockRepo.On("FindByID", mock.Anything, input.ID).Return(nil, domain.ErrNotFound).Once()

		user, err := userUsecase.UpdateProfile(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "User dengan ID tersebut tidak ditemukan")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - Update DB Error", func(t *testing.T) {
		input := domain.UpdateProfileInput{ID: "user-123", Name: "Faridlan Updated"}
		mockUser := &domain.User{ID: input.ID, Name: "Faridlan Lama"}
		dbError := errors.New("failed to save")

		mockRepo.On("FindByID", mock.Anything, input.ID).Return(mockUser, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(dbError).Once()

		user, err := userUsecase.UpdateProfile(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST UPDATE PASSWORD
// ==========================================
func TestUpdatePassword(t *testing.T) {
	mockRepo, userUsecase := setupUserUsecase()

	// Hash password asli agar `bcrypt.Compare` di usecase bisa berjalan sukses
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("passwordLama123"), bcrypt.DefaultCost)

	t.Run("Success", func(t *testing.T) {
		input := domain.UpdatePasswordInput{
			ID:          "user-123",
			OldPassword: "passwordLama123",
			NewPassword: "passwordBaru123",
		}

		mockUser := &domain.User{ID: input.ID, Password: string(hashedPassword)}

		mockRepo.On("FindByID", mock.Anything, input.ID).Return(mockUser, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		err := userUsecase.UpdatePassword(context.Background(), input)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - User Not Found", func(t *testing.T) {
		input := domain.UpdatePasswordInput{
			ID:          "user-123",
			OldPassword: "passwordLama123",
			NewPassword: "passwordBaru123",
		}

		mockRepo.On("FindByID", mock.Anything, input.ID).Return(nil, domain.ErrNotFound).Once()

		err := userUsecase.UpdatePassword(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "User dengan ID tersebut tidak ditemukan")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - Wrong Old Password", func(t *testing.T) {
		input := domain.UpdatePasswordInput{
			ID:          "user-123",
			OldPassword: "passwordSalah", // <-- Sengaja disalahkan
			NewPassword: "passwordBaru123",
		}

		mockUser := &domain.User{ID: input.ID, Password: string(hashedPassword)}

		mockRepo.On("FindByID", mock.Anything, input.ID).Return(mockUser, nil).Once()

		err := userUsecase.UpdatePassword(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Password lama salah")
		assert.ErrorIs(t, err, domain.ErrBadParamInput)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - New Password Same as Old", func(t *testing.T) {
		input := domain.UpdatePasswordInput{
			ID:          "user-123",
			OldPassword: "passwordLama123",
			NewPassword: "passwordLama123", // <-- Sama dengan yang lama
		}

		mockUser := &domain.User{ID: input.ID, Password: string(hashedPassword)}

		mockRepo.On("FindByID", mock.Anything, input.ID).Return(mockUser, nil).Once()

		err := userUsecase.UpdatePassword(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Password baru tidak boleh sama dengan password lama")
		assert.ErrorIs(t, err, domain.ErrBadParamInput)
		mockRepo.AssertExpectations(t)
	})
}
