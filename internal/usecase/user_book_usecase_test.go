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
)

// Helper setup
func setupUserBookUsecase() (*mocks.UserBookRepository, *mocks.BookRepository, domain.UserBookUsecase) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	userBookUsecase := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)
	return mockUserBookRepo, mockBookRepo, userBookUsecase
}

// ==========================================
// TEST TRACK NEW BOOK
// ==========================================
func TestTrackNewBook(t *testing.T) {
	mockUBRepo, mockBookRepo, u := setupUserBookUsecase()
	userID := "user-123"
	bookID := "book-456"

	t.Run("Success", func(t *testing.T) {
		mockBookRepo.On("GetByID", mock.Anything, bookID).Return(&domain.Book{ID: bookID}, nil).Once()
		// Skenario buku belum ada di rak (harus return ErrNotFound agar bisa dilanjut)
		mockUBRepo.On("FindByUserIDAndBookID", mock.Anything, userID, bookID).Return(nil, domain.ErrNotFound).Once()

		mockUBRepo.On("AddBookToShelf", mock.Anything, mock.MatchedBy(func(ub *domain.UserBook) bool {
			return ub.UserID == userID && ub.BookID == bookID && ub.Status == "TO_READ"
		})).Return(nil).Once()

		result, err := u.TrackNewBook(context.Background(), userID, bookID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "TO_READ", result.Status)
		mockBookRepo.AssertExpectations(t)
		mockUBRepo.AssertExpectations(t)
	})

	t.Run("Failed - Master Book Not Found", func(t *testing.T) {
		mockBookRepo.On("GetByID", mock.Anything, bookID).Return(nil, domain.ErrNotFound).Once()

		result, err := u.TrackNewBook(context.Background(), userID, bookID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Buku dengan ID tersebut tidak ditemukan")
		mockBookRepo.AssertExpectations(t)
	})

	t.Run("Failed - Database Error on Checking Master Book", func(t *testing.T) {
		dbError := errors.New("connection reset by peer")
		mockBookRepo.On("GetByID", mock.Anything, bookID).Return(nil, dbError).Once()

		result, err := u.TrackNewBook(context.Background(), userID, bookID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err) // Ini membuktikan bug fix-mu berhasil!
		mockBookRepo.AssertExpectations(t)
	})

	t.Run("Failed - Book Already in Shelf (Conflict)", func(t *testing.T) {
		mockBookRepo.On("GetByID", mock.Anything, bookID).Return(&domain.Book{ID: bookID}, nil).Once()
		// Skenario buku SUDAH ada di rak
		mockUBRepo.On("FindByUserIDAndBookID", mock.Anything, userID, bookID).Return(&domain.UserBookWithMetadata{}, nil).Once()

		result, err := u.TrackNewBook(context.Background(), userID, bookID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrConflict)
		assert.Contains(t, err.Error(), "Buku sudah ada di rak pengguna")
		mockBookRepo.AssertExpectations(t)
		mockUBRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST UPDATE READING STATUS
// ==========================================
func TestUpdateReadingStatus(t *testing.T) {
	mockUBRepo, _, u := setupUserBookUsecase()

	t.Run("Success", func(t *testing.T) {
		input := domain.UpdateUserBookInput{
			UserID: "user-123",
			BookID: "book-456",
			Status: "READING",
			Page:   50,
			Rating: 5,
		}

		existingData := &domain.UserBookWithMetadata{
			UserBook: domain.UserBook{ID: "ub-789", Status: "TO_READ", CurrentPage: 0, Rating: 0},
		}

		mockUBRepo.On("FindByUserIDAndBookID", mock.Anything, input.UserID, input.BookID).Return(existingData, nil).Once()

		mockUBRepo.On("UpdateProgress", mock.Anything, mock.MatchedBy(func(ub *domain.UserBook) bool {
			return ub.Status == "READING" && ub.CurrentPage == 50 && ub.Rating == 5
		})).Return(nil).Once()

		result, err := u.UpdateReadingStatus(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "READING", result.Status)
		mockUBRepo.AssertExpectations(t)
	})

	t.Run("Failed - Shelf Not Found", func(t *testing.T) {
		input := domain.UpdateUserBookInput{UserID: "user-123", BookID: "book-456"}

		mockUBRepo.On("FindByUserIDAndBookID", mock.Anything, input.UserID, input.BookID).Return(nil, domain.ErrNotFound).Once()

		result, err := u.UpdateReadingStatus(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Buku dengan ID tersebut tidak ditemukan")
		mockUBRepo.AssertExpectations(t)
	})

	t.Run("Failed - Database Error on Find", func(t *testing.T) {
		input := domain.UpdateUserBookInput{UserID: "user-123", BookID: "book-456"}
		dbError := errors.New("db down")

		mockUBRepo.On("FindByUserIDAndBookID", mock.Anything, input.UserID, input.BookID).Return(nil, dbError).Once()

		result, err := u.UpdateReadingStatus(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err) // Bukti bug fix-mu jalan
		mockUBRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST GET USER LIBRARY
// ==========================================
func TestGetUserLibrary(t *testing.T) {
	mockUBRepo, _, u := setupUserBookUsecase()

	t.Run("Success", func(t *testing.T) {
		params := domain.PaginationQuery{Page: 1, Limit: 10}
		mockData := []*domain.UserBookWithMetadata{{}, {}}
		var totalItems int64 = 15 // Jika total 15 dan limit 10, maka totalPages harusnya 2

		mockUBRepo.On("FindAllByUserID", mock.Anything, "user-123", "READING", params).Return(mockData, totalItems, nil).Once()

		books, meta, err := u.GetUserLibrary(context.Background(), "user-123", "READING", params)

		assert.NoError(t, err)
		assert.Len(t, books, 2)
		assert.Equal(t, 2, meta.TotalPages)
		assert.Equal(t, int64(15), meta.TotalItems)
		mockUBRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST GET USER BOOK DETAIL
// ==========================================
func TestGetUserBookDetail(t *testing.T) {
	mockUBRepo, _, u := setupUserBookUsecase()

	t.Run("Success", func(t *testing.T) {
		mockData := &domain.UserBookWithMetadata{}
		mockUBRepo.On("GetDetailByID", mock.Anything, "user-123", "book-456").Return(mockData, nil).Once()

		result, err := u.GetUserBookDetail(context.Background(), "user-123", "book-456")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockUBRepo.AssertExpectations(t)
	})

	t.Run("Failed - Not Found", func(t *testing.T) {
		mockUBRepo.On("GetDetailByID", mock.Anything, "user-123", "book-456").Return(nil, domain.ErrNotFound).Once()

		result, err := u.GetUserBookDetail(context.Background(), "user-123", "book-456")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Buku dengan ID tersebut tidak ditemukan")
		mockUBRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST DELETE BOOK FROM SHELF
// ==========================================
func TestDeleteBookFromShelf(t *testing.T) {
	mockUBRepo, _, u := setupUserBookUsecase()

	t.Run("Success", func(t *testing.T) {
		mockUBRepo.On("FindByUserIDAndBookID", mock.Anything, "user-123", "book-456").Return(&domain.UserBookWithMetadata{}, nil).Once()
		mockUBRepo.On("Delete", mock.Anything, "user-123", "book-456").Return(nil).Once()

		err := u.DeleteBookFromShelf(context.Background(), "user-123", "book-456")

		assert.NoError(t, err)
		mockUBRepo.AssertExpectations(t)
	})

	t.Run("Failed - Not Found", func(t *testing.T) {
		mockUBRepo.On("FindByUserIDAndBookID", mock.Anything, "user-123", "book-456").Return(nil, domain.ErrNotFound).Once()

		err := u.DeleteBookFromShelf(context.Background(), "user-123", "book-456")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Buku dengan ID tersebut tidak ditemukan")
		mockUBRepo.AssertExpectations(t)
	})

	t.Run("Failed - Database Error on Find", func(t *testing.T) {
		dbError := errors.New("timeout")
		mockUBRepo.On("FindByUserIDAndBookID", mock.Anything, "user-123", "book-456").Return(nil, dbError).Once()

		err := u.DeleteBookFromShelf(context.Background(), "user-123", "book-456")

		assert.Error(t, err)
		assert.Equal(t, dbError, err) // Terproteksi dari bug
		mockUBRepo.AssertExpectations(t)
	})
}
