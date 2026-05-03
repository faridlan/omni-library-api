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

func TestFetchAndSaveMetadata_BukuAdaDiLokal(t *testing.T) {

	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)

	bookUC := usecase.NewBookUsecase(mockRepo, mockFetcher)

	isbnTest := "9781234567890"
	dummyBook := &domain.Book{
		ID:    "buku-001",
		ISBN:  isbnTest,
		Title: "Golang Clean Architecture",
	}

	mockRepo.On("GetByISBN", mock.Anything, isbnTest).Return(dummyBook, nil)

	result, err := bookUC.FetchAndSaveMetadata(context.Background(), isbnTest)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dummyBook.Title, result.Title)

	mockFetcher.AssertNotCalled(t, "FetchByISBN", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestFetchAndSaveMetadata_BukuTidakAdaDiLokal_AmbilDariAPI(t *testing.T) {

	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	bookUC := usecase.NewBookUsecase(mockRepo, mockFetcher)

	isbnTest := "9780134685991"
	fetchedBook := &domain.Book{
		ID:    "buku-baru-002",
		ISBN:  isbnTest,
		Title: "Effective Java",
	}

	mockRepo.On("GetByISBN", mock.Anything, isbnTest).Return(nil, nil)

	mockFetcher.On("FetchByISBN", mock.Anything, isbnTest).Return(fetchedBook, nil)

	mockRepo.On("Create", mock.Anything, fetchedBook).Return(nil)

	result, err := bookUC.FetchAndSaveMetadata(context.Background(), isbnTest)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, fetchedBook.Title, result.Title)

	mockRepo.AssertExpectations(t)
	mockFetcher.AssertExpectations(t)
}

func TestFetchAndSaveMetadata_BukuTidakDitemukanDiManapun(t *testing.T) {

	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	bookUC := usecase.NewBookUsecase(mockRepo, mockFetcher)

	isbnTest := "9999999999999"

	mockRepo.On("GetByISBN", mock.Anything, isbnTest).Return(nil, nil)

	mockFetcher.On("FetchByISBN", mock.Anything, isbnTest).Return(nil, nil)

	result, err := bookUC.FetchAndSaveMetadata(context.Background(), isbnTest)

	assert.Error(t, err)
	assert.Nil(t, result)

	assert.ErrorIs(t, err, domain.ErrNotFound)

	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateManual_ConflictISBN(t *testing.T) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	reqBook := &domain.Book{ISBN: "123", Title: "Buku Test"}

	mockRepo.On("GetByISBN", mock.Anything, "123").Return(&domain.Book{ID: "old-id"}, nil)

	res, err := uc.CreateManual(context.Background(), reqBook)

	assert.ErrorIs(t, err, domain.ErrConflict)
	assert.Nil(t, res)
}

func TestCreateManual_Success(t *testing.T) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	reqBook := &domain.Book{ISBN: "123", Title: "Buku Test"}

	// Naskah: ISBN belum ada, lalu Create sukses
	mockRepo.On("GetByISBN", mock.Anything, "123").Return(nil, nil)
	mockRepo.On("Create", mock.Anything, reqBook).Return(nil)

	res, err := uc.CreateManual(context.Background(), reqBook)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Buku Test", res.Title)
}

func TestUpdateBook_NotFound(t *testing.T) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	mockRepo.On("GetByID", mock.Anything, "id-ngawur").Return(nil, nil)

	res, err := uc.UpdateBook(context.Background(), "id-ngawur", &domain.Book{Title: "Baru"})

	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, res)
}

func TestUpdateBook_Success(t *testing.T) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	oldBook := &domain.Book{ID: "id-123", Title: "Lama"}
	reqBook := &domain.Book{Title: "Baru"}

	mockRepo.On("GetByID", mock.Anything, "id-123").Return(oldBook, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Book")).Return(nil)

	res, err := uc.UpdateBook(context.Background(), "id-123", reqBook)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Baru", res.Title)
}

func TestDeleteBook_NotFound(t *testing.T) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	mockRepo.On("GetByID", mock.Anything, "id-ngawur").Return(nil, nil)

	err := uc.DeleteBook(context.Background(), "id-ngawur")

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDeleteBook_Success(t *testing.T) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	existingBook := &domain.Book{ID: "id-123", Title: "Buku Sampah"}

	// Naskah: Buku ketemu, eksekusi hapus berhasil
	mockRepo.On("GetByID", mock.Anything, "id-123").Return(existingBook, nil)
	mockRepo.On("Delete", mock.Anything, "id-123").Return(nil)

	err := uc.DeleteBook(context.Background(), "id-123")

	assert.NoError(t, err)
}

func TestGetAllBooks_Sukses_MenghitungPaginasi(t *testing.T) {

	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	params := domain.PaginationQuery{
		Page:  1,
		Limit: 10,
	}

	mockBooks := []*domain.Book{
		{ID: "book-1", Title: "Buku Golang 101"},
		{ID: "book-2", Title: "Clean Architecture"},
	}

	var totalDataFromDB int64 = 25

	mockRepo.On("GetAll", mock.Anything, params).Return(mockBooks, totalDataFromDB, nil)

	books, meta, err := uc.GetAllBooks(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, books)
	assert.Len(t, books, 2)

	assert.Equal(t, int64(25), meta.TotalItems)
	assert.Equal(t, 10, meta.Limit)
	assert.Equal(t, 1, meta.CurrentPage)

	assert.Equal(t, 3, meta.TotalPages)
}

func TestGetAllBooks_Gagal_DariRepository(t *testing.T) {

	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	params := domain.PaginationQuery{Page: 1, Limit: 10}

	dbError := domain.ErrInternalServerError

	mockRepo.On("GetAll", mock.Anything, params).Return(nil, int64(0), dbError)

	books, meta, err := uc.GetAllBooks(context.Background(), params)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	assert.Nil(t, books)
	assert.Equal(t, 0, meta.TotalPages)
}
