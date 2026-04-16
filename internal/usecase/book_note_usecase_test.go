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

// SKENARIO 1: Gagal karena Buku tidak ada di Rak User
func TestAddNote_BukuTidakAdaDiRak(t *testing.T) {
	mockNoteRepo := new(mocks.BookNoteRepository)
	mockUserBookRepo := new(mocks.UserBookRepository)
	uc := usecase.NewBookNoteUsecase(mockNoteRepo, mockUserBookRepo)

	dummyNote := &domain.BookNote{
		UserBookID: "rak-001",
		Quote:      "Clean Code is a reader-focused development.",
	}

	// NASKAH: Saat Otak mengecek ke rak, ternyata kosong (nil)
	mockUserBookRepo.On("GetByID", mock.Anything, dummyNote.UserBookID).Return(nil, nil)

	// ACTION!
	err := uc.AddNote(context.Background(), dummyNote)

	// VALIDASI: Harus gagal dengan pesan ErrNotFound
	assert.ErrorIs(t, err, domain.ErrNotFound)

	// VALIDASI DISIPLIN: Jangan pernah simpan note kalau bukunya nggak ada!
	mockNoteRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// SKENARIO 2: Sukses Menambahkan Catatan
func TestAddNote_Sukses(t *testing.T) {
	mockNoteRepo := new(mocks.BookNoteRepository)
	mockUserBookRepo := new(mocks.UserBookRepository)
	uc := usecase.NewBookNoteUsecase(mockNoteRepo, mockUserBookRepo)

	dummyNote := &domain.BookNote{
		UserBookID: "rak-001",
		Quote:      "Testing is not a phase, it's a lifestyle.",
	}

	dummyUserBook := &domain.UserBook{ID: "rak-001"}

	// NASKAH A: Bukunya ada di rak
	mockUserBookRepo.On("GetByID", mock.Anything, dummyNote.UserBookID).Return(dummyUserBook, nil)

	// NASKAH B: Proses save/create berjalan lancar
	mockNoteRepo.On("Create", mock.Anything, dummyNote).Return(nil)

	// ACTION!
	err := uc.AddNote(context.Background(), dummyNote)

	// VALIDASI: Tidak boleh ada error
	assert.NoError(t, err)

	mockUserBookRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}
