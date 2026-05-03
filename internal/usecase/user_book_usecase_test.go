package usecase_test

import (
	"context"
	"testing"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/domain/mocks"
	"github.com/faridlan/omni-library-api/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTrackNewBook_BukuMasterTidakAda(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-salah-404"

	mockBookRepo.On("GetByID", mock.Anything, bookID).Return(nil, nil)

	result, err := uc.TrackNewBook(context.Background(), userID, bookID)

	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)

	mockUserBookRepo.AssertNotCalled(t, "GetByUserAndBookID", mock.Anything, mock.Anything, mock.Anything)
}

func TestTrackNewBook_BukuSudahAdaDiRak(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-valid-001"
	dummyMasterBook := &domain.Book{ID: bookID}

	existingShelf := &domain.UserBookWithMetadata{
		UserBook: domain.UserBook{
			UserID: userID,
			BookID: bookID}}

	mockBookRepo.On("GetByID", mock.Anything, bookID).Return(dummyMasterBook, nil)

	mockUserBookRepo.On("GetByBookID", mock.Anything, userID, bookID).Return(existingShelf, nil)

	result, err := uc.TrackNewBook(context.Background(), userID, bookID)

	assert.ErrorIs(t, err, domain.ErrConflict)
	assert.Nil(t, result)

	mockUserBookRepo.AssertNotCalled(t, "AddBookToShelf", mock.Anything, mock.Anything)
}

func TestTrackNewBook_Sukses(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-valid-001"
	dummyMasterBook := &domain.Book{ID: bookID}

	mockBookRepo.On("GetByID", mock.Anything, bookID).Return(dummyMasterBook, nil)

	mockUserBookRepo.On("GetByBookID", mock.Anything, userID, bookID).Return(nil, nil)

	mockUserBookRepo.On("AddBookToShelf", mock.Anything, mock.AnythingOfType("*domain.UserBook")).Return(nil)

	result, err := uc.TrackNewBook(context.Background(), userID, bookID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "TO_READ", result.Status)

	mockBookRepo.AssertExpectations(t)
	mockUserBookRepo.AssertExpectations(t)
}

func TestUpdateReadingStatus_BukuTidakAdaDiRak(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-ngasal-404"

	mockUserBookRepo.On("GetByUserAndBookID", mock.Anything, userID, bookID).Return(nil, nil)

	result, err := uc.UpdateReadingStatus(context.Background(), userID, bookID, "READING", 50, 4)

	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)

	mockUserBookRepo.AssertNotCalled(t, "UpdateProgress", mock.Anything, mock.Anything)
}

func TestUpdateReadingStatus_Sukses(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-valid-001"

	existingTrack := &domain.UserBookWithMetadata{
		UserBook: domain.UserBook{
			UserID:      userID,
			BookID:      bookID,
			Status:      "TO_READ",
			CurrentPage: 0,
			Rating:      0,
		},
	}

	mockUserBookRepo.On("GetByUserAndBookID", mock.Anything, userID, bookID).Return(existingTrack, nil)

	mockUserBookRepo.On("UpdateProgress", mock.Anything, mock.AnythingOfType("*domain.UserBook")).Return(nil)

	result, err := uc.UpdateReadingStatus(context.Background(), userID, bookID, "READING", 125, 5)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "READING", result.Status)
	assert.Equal(t, 125, result.CurrentPage)
	assert.Equal(t, 5, result.Rating)

	mockUserBookRepo.AssertExpectations(t)
}

func TestGetUserLibrary_Sukses_HitungPaginasi(t *testing.T) {

	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	statusFilter := "READING"
	params := domain.PaginationQuery{Page: 1, Limit: 5}

	mockData := []*domain.UserBookWithMetadata{
		{UserBook: domain.UserBook{ID: "ub-1", BookID: "book-1"}},
		{UserBook: domain.UserBook{ID: "ub-2", BookID: "book-2"}},
	}

	var totalData int64 = 12

	mockUserBookRepo.On("GetByUserID", mock.Anything, userID, statusFilter, params).
		Return(mockData, totalData, nil)

	result, meta, err := uc.GetUserLibrary(context.Background(), userID, statusFilter, params)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, int64(12), meta.TotalItems)
	assert.Equal(t, 5, meta.Limit)

	assert.Equal(t, 3, meta.TotalPages)
}

func TestGetUserLibrary_Gagal_DariRepo(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	params := domain.PaginationQuery{Page: 1, Limit: 10}
	expectedErr := domain.ErrInternalServerError

	mockUserBookRepo.On("GetByUserID", mock.Anything, userID, "", params).
		Return(nil, int64(0), expectedErr)

	result, meta, err := uc.GetUserLibrary(context.Background(), userID, "", params)

	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
	assert.Equal(t, 0, meta.TotalPages)
}
