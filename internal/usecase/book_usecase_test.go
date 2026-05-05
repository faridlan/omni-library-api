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

func setupBookUsecase() (*mocks.BookRepository, *mocks.BookMetadataFetcher, domain.BookUsecase) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	u := usecase.NewBookUsecase(mockRepo, mockFetcher)
	return mockRepo, mockFetcher, u
}

// ==========================================
// TEST FETCH AND SAVE METADATA
// ==========================================
func TestFetchAndSaveMetadata(t *testing.T) {
	mockRepo, mockFetcher, u := setupBookUsecase()
	isbn := "9781234567890"

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByISBN", mock.Anything, isbn).Return(nil, domain.ErrNotFound).Once()

		newBook := &domain.Book{ISBN: isbn, Title: "Test Book"}
		mockFetcher.On("FetchByISBN", mock.Anything, isbn).Return(newBook, nil).Once()
		mockRepo.On("Create", mock.Anything, newBook).Return(nil).Once()

		result, err := u.FetchAndSaveMetadata(context.Background(), isbn)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Book", result.Title)
		mockRepo.AssertExpectations(t)
		mockFetcher.AssertExpectations(t)
	})

	t.Run("Failed - Book Already Exists (Conflict)", func(t *testing.T) {
		existingBook := &domain.Book{ISBN: isbn}
		mockRepo.On("GetByISBN", mock.Anything, isbn).Return(existingBook, nil).Once()

		result, err := u.FetchAndSaveMetadata(context.Background(), isbn)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrConflict)
		assert.Contains(t, err.Error(), "Buku dengan ISBN tersebut sudah ada di sistem")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - Not Found in Google Books", func(t *testing.T) {
		mockRepo.On("GetByISBN", mock.Anything, isbn).Return(nil, domain.ErrNotFound).Once()
		mockFetcher.On("FetchByISBN", mock.Anything, isbn).Return(nil, nil).Once() // Asumsi fetcher return nil jika tidak ketemu

		result, err := u.FetchAndSaveMetadata(context.Background(), isbn)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Buku dengan ISBN tersebut tidak ditemukan")
		mockRepo.AssertExpectations(t)
		mockFetcher.AssertExpectations(t)
	})
}

// ==========================================
// TEST GET ALL BOOKS
// ==========================================
func TestGetAllBooks(t *testing.T) {
	mockRepo, _, u := setupBookUsecase()

	t.Run("Success", func(t *testing.T) {
		params := domain.PaginationQuery{Page: 1, Limit: 10}
		mockData := []*domain.Book{{}, {}}
		var totalItems int64 = 25 // 25 items, limit 10 = 3 pages

		mockRepo.On("GetAll", mock.Anything, params).Return(mockData, totalItems, nil).Once()

		books, meta, err := u.GetAllBooks(context.Background(), params)

		assert.NoError(t, err)
		assert.Len(t, books, 2)
		assert.Equal(t, 3, meta.TotalPages)
		assert.Equal(t, int64(25), meta.TotalItems)
		mockRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST GET BOOK BY ID
// ==========================================
func TestGetBookByID(t *testing.T) {
	mockRepo, _, u := setupBookUsecase()

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, "book-123").Return(&domain.Book{ID: "book-123"}, nil).Once()

		result, err := u.GetBookByID(context.Background(), "book-123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "book-123", result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, "book-123").Return(nil, domain.ErrNotFound).Once()

		result, err := u.GetBookByID(context.Background(), "book-123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Buku dengan ID tersebut tidak ditemukan")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - Database Error", func(t *testing.T) {
		dbError := errors.New("db error")
		mockRepo.On("GetByID", mock.Anything, "book-123").Return(nil, dbError).Once()

		result, err := u.GetBookByID(context.Background(), "book-123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err) // Ini akan gagal jika kamu belum menambakan `return nil, err` di kodemu!
		mockRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST CREATE MANUAL
// ==========================================
func TestCreateManual(t *testing.T) {
	mockRepo, _, u := setupBookUsecase()

	t.Run("Success", func(t *testing.T) {
		input := domain.CreateBookInput{
			ISBN:  "12345",
			Title: "Manual Book",
		}

		mockRepo.On("GetByISBN", mock.Anything, "12345").Return(nil, domain.ErrNotFound).Once()
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.Book) bool {
			return b.Title == "Manual Book" && b.ISBN == "12345"
		})).Return(nil).Once()

		result, err := u.CreateManual(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed - ISBN Conflict", func(t *testing.T) {
		input := domain.CreateBookInput{ISBN: "12345"}
		mockRepo.On("GetByISBN", mock.Anything, "12345").Return(&domain.Book{}, nil).Once()

		result, err := u.CreateManual(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrConflict)
		assert.Contains(t, err.Error(), "Buku dengan ISBN tersebut sudah terdaftar")
		mockRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST UPDATE BOOK
// ==========================================
func TestUpdateBook(t *testing.T) {
	mockRepo, _, u := setupBookUsecase()

	t.Run("Success", func(t *testing.T) {
		input := domain.UpdateBookInput{
			ID:    "book-123",
			Title: "Updated Title",
		}
		existingBook := &domain.Book{ID: "book-123", Title: "Old Title"}

		mockRepo.On("GetByID", mock.Anything, "book-123").Return(existingBook, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(b *domain.Book) bool {
			return b.Title == "Updated Title"
		})).Return(nil).Once()

		result, err := u.UpdateBook(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated Title", result.Title)
		mockRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST DELETE BOOK
// ==========================================
func TestDeleteBook(t *testing.T) {
	mockRepo, _, u := setupBookUsecase()

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, "book-123").Return(&domain.Book{}, nil).Once()
		mockRepo.On("Delete", mock.Anything, "book-123").Return(nil).Once()

		err := u.DeleteBook(context.Background(), "book-123")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
