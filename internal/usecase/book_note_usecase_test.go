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

func TestAddNote_BukuTidakAdaDiRak(t *testing.T) {
	mockNoteRepo := new(mocks.BookNoteRepository)
	mockUserBookRepo := new(mocks.UserBookRepository)
	uc := usecase.NewBookNoteUsecase(mockNoteRepo, mockUserBookRepo)

	dummyNote := &domain.BookNote{
		UserBookID: "rak-001",
		Quote:      "Clean Code is a reader-focused development.",
	}

	mockUserBookRepo.On("GetByID", mock.Anything, dummyNote.UserBookID).Return(nil, nil)

	err := uc.AddNote(context.Background(), dummyNote)

	assert.ErrorIs(t, err, domain.ErrNotFound)

	mockNoteRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAddNote_Sukses(t *testing.T) {
	mockNoteRepo := new(mocks.BookNoteRepository)
	mockUserBookRepo := new(mocks.UserBookRepository)
	uc := usecase.NewBookNoteUsecase(mockNoteRepo, mockUserBookRepo)

	dummyNote := &domain.BookNote{
		UserBookID: "rak-001",
		Quote:      "Testing is not a phase, it's a lifestyle.",
	}

	dummyUserBook := &domain.UserBook{ID: "rak-001"}

	mockUserBookRepo.On("GetByID", mock.Anything, dummyNote.UserBookID).Return(dummyUserBook, nil)

	mockNoteRepo.On("Create", mock.Anything, dummyNote).Return(nil)

	err := uc.AddNote(context.Background(), dummyNote)

	assert.NoError(t, err)

	mockUserBookRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestGetNotesByUserBookID_Sukses_HitungPaginasi(t *testing.T) {
	mockNoteRepo := new(mocks.BookNoteRepository)
	mockUserBookRepo := new(mocks.UserBookRepository)
	uc := usecase.NewBookNoteUsecase(mockNoteRepo, mockUserBookRepo)

	userBookID := "rak-123"
	params := domain.PaginationQuery{Page: 1, Limit: 3}

	dummyNotes := []*domain.BookNote{
		{ID: "note-1", Quote: "Quote 1"},
		{ID: "note-2", Quote: "Quote 2"},
		{ID: "note-3", Quote: "Quote 3"},
	}

	var totalData int64 = 7

	mockUserBookRepo.On("GetByID", mock.Anything, userBookID).Return(&domain.UserBook{ID: userBookID}, nil)

	mockNoteRepo.On("GetByUserBookID", mock.Anything, userBookID, params).Return(dummyNotes, totalData, nil)

	result, meta, err := uc.GetNotesForBook(context.Background(), userBookID, params)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	assert.Equal(t, int64(7), meta.TotalItems)
	assert.Equal(t, 3, meta.Limit)

	assert.Equal(t, 3, meta.TotalPages)
}

func TestGetNotesByUserBookID_BukuTidakDitemukan(t *testing.T) {

	mockNoteRepo := new(mocks.BookNoteRepository)
	mockUserBookRepo := new(mocks.UserBookRepository)
	uc := usecase.NewBookNoteUsecase(mockNoteRepo, mockUserBookRepo)

	userBookID := "rak-ngawur"
	params := domain.PaginationQuery{Page: 1, Limit: 10}

	mockUserBookRepo.On("GetByID", mock.Anything, userBookID).Return(nil, nil)

	result, meta, err := uc.GetNotesForBook(context.Background(), userBookID, params)

	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)
	assert.Equal(t, 0, meta.TotalPages)

	mockNoteRepo.AssertNotCalled(t, "GetByUserBookID", mock.Anything, mock.Anything, mock.Anything)
}
