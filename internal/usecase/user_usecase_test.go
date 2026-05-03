package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/faridlan/omni-library-api/internal/delivery/http/dto"
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/domain/mocks"
	"github.com/faridlan/omni-library-api/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func setupUserUsecase() (*mocks.UserRepository, domain.UserUsecase) {
	mockRepo := new(mocks.UserRepository)
	userUsecase := usecase.NewUserUsecase(mockRepo)
	return mockRepo, userUsecase
}

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
		mockRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("db error")).Once()

		user, err := userUsecase.GetProfile(context.Background(), userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "data tidak ditemukan", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateProfile(t *testing.T) {
	mockRepo, userUsecase := setupUserUsecase()
	userID := "user-123"

	t.Run("Success", func(t *testing.T) {
		req := &dto.UpdateProfileRequest{Name: "Faridlan Updated"}
		mockUser := &domain.User{ID: userID, Name: "Faridlan Lama"}

		mockRepo.On("FindByID", mock.Anything, userID).Return(mockUser, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.Name == "Faridlan Updated"
		})).Return(nil).Once()

		user, err := userUsecase.UpdateProfile(context.Background(), userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Faridlan Updated", user.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - User Not Found", func(t *testing.T) {
		req := &dto.UpdateProfileRequest{Name: "Faridlan Updated"}

		mockRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("db error")).Once()

		user, err := userUsecase.UpdateProfile(context.Background(), userID, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "data tidak ditemukan", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdatePassword(t *testing.T) {
	mockRepo, userUsecase := setupUserUsecase()
	userID := "user-123"

	// Bikin hash password asli untuk testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("passwordLama123"), bcrypt.DefaultCost)

	t.Run("Success", func(t *testing.T) {
		req := &dto.UpdatePasswordRequest{
			OldPassword:     "passwordLama123",
			NewPassword:     "passwordBaru123",
			ConfirmPassword: "passwordBaru123",
		}

		mockUser := &domain.User{ID: userID, Password: string(hashedPassword)}

		mockRepo.On("FindByID", mock.Anything, userID).Return(mockUser, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		err := userUsecase.UpdatePassword(context.Background(), userID, req)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - Old Password Wrong", func(t *testing.T) {
		req := &dto.UpdatePasswordRequest{
			OldPassword:     "passwordSalah",
			NewPassword:     "passwordBaru123",
			ConfirmPassword: "passwordBaru123",
		}

		mockUser := &domain.User{ID: userID, Password: string(hashedPassword)}

		mockRepo.On("FindByID", mock.Anything, userID).Return(mockUser, nil).Once()

		err := userUsecase.UpdatePassword(context.Background(), userID, req)

		assert.Error(t, err)
		assert.Equal(t, "parameter atau format data tidak valid", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - New Password Same as Old", func(t *testing.T) {
		req := &dto.UpdatePasswordRequest{
			OldPassword:     "passwordLama123",
			NewPassword:     "passwordLama123", // Sama dengan yang lama
			ConfirmPassword: "passwordLama123",
		}

		mockUser := &domain.User{ID: userID, Password: string(hashedPassword)}

		mockRepo.On("FindByID", mock.Anything, userID).Return(mockUser, nil).Once()

		err := userUsecase.UpdatePassword(context.Background(), userID, req)

		assert.Error(t, err)
		assert.Equal(t, "parameter atau format data tidak valid", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - Validation Error (Confirm Mismatch)", func(t *testing.T) {
		req := &dto.UpdatePasswordRequest{
			OldPassword:     "passwordLama123",
			NewPassword:     "passwordBaru123",
			ConfirmPassword: "bedaPassword", // Tidak sama
		}

		// Karena validasi struct (utils.ValidateStruct) berjalan di awal, FindByID tidak akan pernah dipanggil
		err := userUsecase.UpdatePassword(context.Background(), userID, req)

		assert.Error(t, err)
		// Pesan error di sini tergantung implementasi utils.ValidateStruct-mu
	})
}
