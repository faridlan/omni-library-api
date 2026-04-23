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

// ==========================================
// TEST GET NOTES BY USER BOOK (PAGINATION)
// ==========================================

// SKENARIO 3: Sukses Mengambil Daftar Catatan dengan Paginasi
func TestGetNotesByUserBookID_Sukses_HitungPaginasi(t *testing.T) {
	// 1. SIAPKAN STUNTMAN
	mockNoteRepo := new(mocks.BookNoteRepository)
	mockUserBookRepo := new(mocks.UserBookRepository)
	uc := usecase.NewBookNoteUsecase(mockNoteRepo, mockUserBookRepo)

	// 2. SIAPKAN DATA PALSU & PARAMETER
	userBookID := "rak-123"
	params := domain.PaginationQuery{Page: 1, Limit: 3} // Limit 3 per halaman

	dummyNotes := []*domain.BookNote{
		{ID: "note-1", Quote: "Quote 1"},
		{ID: "note-2", Quote: "Quote 2"},
		{ID: "note-3", Quote: "Quote 3"},
	}

	// Skenario: Total ada 7 catatan di database untuk buku ini
	var totalData int64 = 7

	// 3. ATUR NASKAH (Ekspektasi)
	// Naskah A: Pastikan buku memang ada di rak user
	mockUserBookRepo.On("GetByID", mock.Anything, userBookID).Return(&domain.UserBook{ID: userBookID}, nil)

	// Naskah B: Repo mengembalikan 3 data dan total items 7
	mockNoteRepo.On("GetByUserBookID", mock.Anything, userBookID, params).Return(dummyNotes, totalData, nil)

	// 4. ACTION!
	result, meta, err := uc.GetNotesForBook(context.Background(), userBookID, params)

	// 5. VALIDASI HASIL
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// 6. VALIDASI MATEMATIKA PAGINASI (The Magic of math.Ceil)
	assert.Equal(t, int64(7), meta.TotalItems)
	assert.Equal(t, 3, meta.Limit)

	// 7 data / 3 per halaman = 2.33 -> Harus jadi 3 Halaman
	assert.Equal(t, 3, meta.TotalPages)
}

// SKENARIO 4: Gagal mengambil catatan karena Bukunya tidak ditemukan
func TestGetNotesByUserBookID_BukuTidakDitemukan(t *testing.T) {
	// 1. SIAPKAN STUNTMAN
	mockNoteRepo := new(mocks.BookNoteRepository)
	mockUserBookRepo := new(mocks.UserBookRepository)
	uc := usecase.NewBookNoteUsecase(mockNoteRepo, mockUserBookRepo)

	userBookID := "rak-ngawur"
	params := domain.PaginationQuery{Page: 1, Limit: 10}

	// 2. NASKAH: Rak user kosong
	mockUserBookRepo.On("GetByID", mock.Anything, userBookID).Return(nil, nil)

	// 3. ACTION!
	result, meta, err := uc.GetNotesForBook(context.Background(), userBookID, params)

	// 4. VALIDASI
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)
	assert.Equal(t, 0, meta.TotalPages)

	// Pastikan Repo Note tidak dipanggil kalau bukunya saja tidak ketemu
	mockNoteRepo.AssertNotCalled(t, "GetByUserBookID", mock.Anything, mock.Anything, mock.Anything)
}
