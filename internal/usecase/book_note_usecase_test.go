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
func setupBookNoteUsecase() (*mocks.BookNoteRepository, *mocks.UserBookRepository, domain.BookNoteUsecase) {
	mockNoteRepo := new(mocks.BookNoteRepository)
	mockUBRepo := new(mocks.UserBookRepository)
	u := usecase.NewBookNoteUsecase(mockNoteRepo, mockUBRepo)
	return mockNoteRepo, mockUBRepo, u
}

// ==========================================
// TEST ADD NOTE
// ==========================================
func TestAddNote(t *testing.T) {
	mockNoteRepo, mockUBRepo, u := setupBookNoteUsecase()
	input := domain.CreateBookNoteInput{
		UserBookID:    "ub-123",
		Quote:         "Bekerjalah seperti programmer pemalas",
		PageReference: 42,
		Tags:          []string{"Inspiratif"},
	}

	t.Run("Success", func(t *testing.T) {
		mockUBRepo.On("FindByID", mock.Anything, input.UserBookID).Return(&domain.UserBook{}, nil).Once()
		mockNoteRepo.On("Create", mock.Anything, mock.MatchedBy(func(n *domain.BookNote) bool {
			return n.Quote == input.Quote && n.UserBookID == input.UserBookID
		})).Return(nil).Once()

		result, err := u.AddNote(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, input.Quote, result.Quote)
		mockUBRepo.AssertExpectations(t)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Failed - UserBook Not Found", func(t *testing.T) {
		mockUBRepo.On("FindByID", mock.Anything, input.UserBookID).Return(nil, domain.ErrNotFound).Once()

		result, err := u.AddNote(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Rak buku tidak ditemukan")
		mockUBRepo.AssertExpectations(t)
	})

	t.Run("Failed - UserBook DB Error", func(t *testing.T) {
		dbErr := errors.New("db error")
		mockUBRepo.On("FindByID", mock.Anything, input.UserBookID).Return(nil, dbErr).Once()

		result, err := u.AddNote(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbErr, err) // Bukti bug fix berhasil
		mockUBRepo.AssertExpectations(t)
	})

	t.Run("Failed - Create DB Error", func(t *testing.T) {
		mockUBRepo.On("FindByID", mock.Anything, input.UserBookID).Return(&domain.UserBook{}, nil).Once()
		mockNoteRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.BookNote")).Return(errors.New("insert failed")).Once()

		result, err := u.AddNote(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrInternalServerError, err) // Diubah oleh usecase menjadi InternalServerError
		mockUBRepo.AssertExpectations(t)
		mockNoteRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST GET NOTES FOR BOOK
// ==========================================
func TestGetNotesForBook(t *testing.T) {
	mockNoteRepo, mockUBRepo, u := setupBookNoteUsecase()
	ubID := "ub-123"
	params := domain.PaginationQuery{Page: 1, Limit: 10}

	t.Run("Success", func(t *testing.T) {
		mockUBRepo.On("FindByID", mock.Anything, ubID).Return(&domain.UserBook{}, nil).Once()

		mockData := []*domain.BookNote{{}, {}, {}}
		var totalItems int64 = 25 // Limit 10 -> Total Pages: 3
		mockNoteRepo.On("FindAllByUserBookID", mock.Anything, ubID, params).Return(mockData, totalItems, nil).Once()

		notes, meta, err := u.GetNotesForBook(context.Background(), ubID, params)

		assert.NoError(t, err)
		assert.Len(t, notes, 3)
		assert.Equal(t, 3, meta.TotalPages)
		assert.Equal(t, int64(25), meta.TotalItems)
		mockUBRepo.AssertExpectations(t)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Failed - UserBook Not Found", func(t *testing.T) {
		mockUBRepo.On("FindByID", mock.Anything, ubID).Return(nil, domain.ErrNotFound).Once()

		notes, _, err := u.GetNotesForBook(context.Background(), ubID, params)

		assert.Error(t, err)
		assert.Nil(t, notes)
		assert.Contains(t, err.Error(), "Rak buku tidak ditemukan")
		mockUBRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST DELETE NOTE
// ==========================================
func TestDeleteNote(t *testing.T) {
	mockNoteRepo, _, u := setupBookNoteUsecase()
	noteID := "note-123"

	t.Run("Success", func(t *testing.T) {
		mockNoteRepo.On("FindByID", mock.Anything, noteID).Return(&domain.BookNote{}, nil).Once()
		mockNoteRepo.On("Delete", mock.Anything, noteID).Return(nil).Once()

		err := u.DeleteNote(context.Background(), noteID)

		assert.NoError(t, err)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Failed - Note Not Found", func(t *testing.T) {
		mockNoteRepo.On("FindByID", mock.Anything, noteID).Return(nil, domain.ErrNotFound).Once()

		err := u.DeleteNote(context.Background(), noteID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Catatan tidak ditemukan")
		mockNoteRepo.AssertExpectations(t)
	})
}

// ==========================================
// TEST UPDATE NOTE
// ==========================================
func TestUpdateNote(t *testing.T) {
	mockNoteRepo, _, u := setupBookNoteUsecase()
	input := domain.UpdateBookNoteInput{
		ID:            "note-123",
		Quote:         "Updated Quote",
		PageReference: 100,
		Tags:          []string{"Updated"},
	}

	t.Run("Success", func(t *testing.T) {
		existingNote := &domain.BookNote{ID: "note-123", Quote: "Old Quote"}
		mockNoteRepo.On("FindByID", mock.Anything, input.ID).Return(existingNote, nil).Once()

		mockNoteRepo.On("Update", mock.Anything, mock.MatchedBy(func(n *domain.BookNote) bool {
			return n.Quote == "Updated Quote" && n.PageReference == 100
		})).Return(nil).Once()

		result, err := u.UpdateNote(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated Quote", result.Quote)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Failed - Note Not Found", func(t *testing.T) {
		mockNoteRepo.On("FindByID", mock.Anything, input.ID).Return(nil, domain.ErrNotFound).Once()

		result, err := u.UpdateNote(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Catatan tidak ditemukan")
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Failed - Update DB Error", func(t *testing.T) {
		existingNote := &domain.BookNote{ID: "note-123"}
		mockNoteRepo.On("FindByID", mock.Anything, input.ID).Return(existingNote, nil).Once()
		mockNoteRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.BookNote")).Return(errors.New("update failed")).Once()

		result, err := u.UpdateNote(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrInternalServerError, err) // Bukti error mapping jalan
		mockNoteRepo.AssertExpectations(t)
	})
}
