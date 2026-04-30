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

// SKENARIO 1: Buku Master Tidak Ditemukan
func TestTrackNewBook_BukuMasterTidakAda(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository) // Butuh stuntman BookRepo karena kita ngecek ke tabel master
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-salah-404"

	// Naskah: Saat Otak nanya ID buku ini ke tabel master, jawab "Nggak ketemu (nil)"
	mockBookRepo.On("GetByID", mock.Anything, bookID).Return(nil, nil)

	// Action!
	result, err := uc.TrackNewBook(context.Background(), userID, bookID)

	// Validasi: Harus error NotFound
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)

	// Validasi Disiplin: Otak TIDAK BOLEH lanjut ngecek ke rak user!
	mockUserBookRepo.AssertNotCalled(t, "GetByUserAndBookID", mock.Anything, mock.Anything, mock.Anything)
}

// SKENARIO 2: Buku Sudah Ada Di Rak User
func TestTrackNewBook_BukuSudahAdaDiRak(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-valid-001"
	dummyMasterBook := &domain.Book{ID: bookID}
	// existingShelf := &domain.UserBook{UserID: userID, BookID: bookID}
	existingShelf := &domain.UserBookWithMetadata{
		UserBook: domain.UserBook{
			UserID: userID,
			BookID: bookID}}

	// Naskah A: Buku master ketemu
	mockBookRepo.On("GetByID", mock.Anything, bookID).Return(dummyMasterBook, nil)

	// Naskah B: Saat ngecek ke rak user, ternyata SUDAH ADA datanya!
	mockUserBookRepo.On("GetByBookID", mock.Anything, userID, bookID).Return(existingShelf, nil)

	// Action!
	result, err := uc.TrackNewBook(context.Background(), userID, bookID)

	// Validasi: Harus error Conflict (409)
	assert.ErrorIs(t, err, domain.ErrConflict)
	assert.Nil(t, result)

	// Validasi Disiplin: Otak TIDAK BOLEH nyuruh save (AddBookToShelf)
	mockUserBookRepo.AssertNotCalled(t, "AddBookToShelf", mock.Anything, mock.Anything)
}

// SKENARIO 3: Jalan Mulus (Happy Path)
func TestTrackNewBook_Sukses(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-valid-001"
	dummyMasterBook := &domain.Book{ID: bookID}

	// Naskah A: Buku master ketemu
	mockBookRepo.On("GetByID", mock.Anything, bookID).Return(dummyMasterBook, nil)

	// Naskah B: Dicek ke rak user, datanya BELUM ADA (nil) -> Aman!
	mockUserBookRepo.On("GetByBookID", mock.Anything, userID, bookID).Return(nil, nil)

	// Naskah C: Otak menyuruh save ke rak. Kita suruh Stuntman pura-pura sukses save.
	// Kita pakai mock.AnythingOfType untuk ngakalin karena kita nggak tahu alamat pointer memorinya secara pasti.
	mockUserBookRepo.On("AddBookToShelf", mock.Anything, mock.AnythingOfType("*domain.UserBook")).Return(nil)

	// Action!
	result, err := uc.TrackNewBook(context.Background(), userID, bookID)

	// Validasi
	assert.NoError(t, err) // Nggak boleh ada error
	assert.NotNil(t, result)
	assert.Equal(t, "TO_READ", result.Status) // Pastikan status otomatis terset "TO_READ"

	// Validasi Kedisiplinan: Semua Stuntman harus menjalankan naskahnya
	mockBookRepo.AssertExpectations(t)
	mockUserBookRepo.AssertExpectations(t)
}

// SKENARIO 4: Update Progres - Gagal Karena Buku Tidak Ada di Rak
func TestUpdateReadingStatus_BukuTidakAdaDiRak(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-ngasal-404"

	// NASKAH: Saat Otak mengecek ke rak user, Stuntman menjawab "Kosong (nil)"
	mockUserBookRepo.On("GetByUserAndBookID", mock.Anything, userID, bookID).Return(nil, nil)

	// ACTION! (Mencoba update ke halaman 50)
	result, err := uc.UpdateReadingStatus(context.Background(), userID, bookID, "READING", 50, 4)

	// VALIDASI
	assert.ErrorIs(t, err, domain.ErrNotFound) // Harus lapor NotFound
	assert.Nil(t, result)

	// VALIDASI DISIPLIN: Karena buku nggak ada, jangan pernah coba-coba manggil fungsi Update!
	mockUserBookRepo.AssertNotCalled(t, "UpdateProgress", mock.Anything, mock.Anything)
}

// SKENARIO 5: Update Progres - Sukses (Happy Path)
func TestUpdateReadingStatus_Sukses(t *testing.T) {
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	userID := "user-123"
	bookID := "buku-valid-001"

	// Anggap saja ini data lama sebelum di-update
	// existingTrack := &domain.UserBook{
	// 	UserID:      userID,
	// 	BookID:      bookID,
	// 	Status:      "TO_READ",
	// 	CurrentPage: 0,
	// 	Rating:      0,
	// }

	existingTrack := &domain.UserBookWithMetadata{
		UserBook: domain.UserBook{
			UserID:      userID,
			BookID:      bookID,
			Status:      "TO_READ",
			CurrentPage: 0,
			Rating:      0,
		},
	}

	// NASKAH A: Saat dicek ke rak, Stuntman memberikan data lama di atas
	mockUserBookRepo.On("GetByUserAndBookID", mock.Anything, userID, bookID).Return(existingTrack, nil)

	// NASKAH B: Saat disuruh simpan update, Stuntman pura-pura sukses
	mockUserBookRepo.On("UpdateProgress", mock.Anything, mock.AnythingOfType("*domain.UserBook")).Return(nil)

	// ACTION! (User membaca sampai halaman 125, ngasih rating 5)
	result, err := uc.UpdateReadingStatus(context.Background(), userID, bookID, "READING", 125, 5)

	// VALIDASI
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Pastikan Otak benar-benar menimpa datanya!
	assert.Equal(t, "READING", result.Status)
	assert.Equal(t, 125, result.CurrentPage)
	assert.Equal(t, 5, result.Rating)

	// Validasi bahwa semua Stuntman menjalankan naskahnya
	mockUserBookRepo.AssertExpectations(t)
}

// ==========================================
// TEST GET USER LIBRARY (PAGINATION)
// ==========================================

func TestGetUserLibrary_Sukses_HitungPaginasi(t *testing.T) {
	// 1. SIAPKAN STUNTMAN
	mockUserBookRepo := new(mocks.UserBookRepository)
	mockBookRepo := new(mocks.BookRepository)
	uc := usecase.NewUserBookUsecase(mockUserBookRepo, mockBookRepo)

	// 2. SIAPKAN DATA PALSU & PARAMETER
	userID := "user-123"
	statusFilter := "READING"
	params := domain.PaginationQuery{Page: 1, Limit: 5} // Limit 5 per halaman

	// Pura-puranya kita balikin 2 buku
	mockData := []*domain.UserBookWithMetadata{
		{UserBook: domain.UserBook{ID: "ub-1", BookID: "book-1"}},
		{UserBook: domain.UserBook{ID: "ub-2", BookID: "book-2"}},
	}

	// Total data di DB ada 12
	var totalData int64 = 12

	// 3. ATUR SKENARIO (Ekspektasi)
	mockUserBookRepo.On("GetByUserID", mock.Anything, userID, statusFilter, params).
		Return(mockData, totalData, nil)

	// 4. EKSEKUSI ACTION!
	result, meta, err := uc.GetUserLibrary(context.Background(), userID, statusFilter, params)

	// 5. VALIDASI HASIL
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// 6. VALIDASI MATEMATIKA PAGINASI
	assert.Equal(t, int64(12), meta.TotalItems)
	assert.Equal(t, 5, meta.Limit)

	// 12 data dibagi 5 per halaman = 2.4 (Dibulatkan ke atas jadi 3 Halaman)
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
