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
	// 1. SIAPKAN STUNTMAN (Mocks)
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)

	// 2. SUNTIKKAN STUNTMAN KE DALAM USECASE
	bookUC := usecase.NewBookUsecase(mockRepo, mockFetcher)

	// 3. SIAPKAN DATA PALSU
	isbnTest := "9781234567890"
	dummyBook := &domain.Book{
		ID:    "buku-001",
		ISBN:  isbnTest,
		Title: "Golang Clean Architecture",
	}

	// 4. ATUR SKENARIO (Ekspektasi)
	// Kita perintahkan Stuntman Repo: "Kalau ada yang manggil fungsi GetByISBN pakai ISBN ini,
	// kembalikan dummyBook dan jangan ada error (nil)!"
	mockRepo.On("GetByISBN", mock.Anything, isbnTest).Return(dummyBook, nil)

	// 5. EKSEKUSI FUNGSI ASLI! (ACTION!)
	result, err := bookUC.FetchAndSaveMetadata(context.Background(), isbnTest)

	// 6. VALIDASI HASILNYA (Assert)
	// Kita pakai library Testify Assert agar mengeceknya mudah bagaikan membaca bahasa Inggris
	assert.NoError(t, err)                         // Pastikan tidak ada error
	assert.NotNil(t, result)                       // Pastikan hasilnya tidak kosong
	assert.Equal(t, dummyBook.Title, result.Title) // Pastikan judulnya sama

	// 7. VALIDASI ATURAN BISNIS (Sangat Penting!)
	// Karena buku sudah ada di lokal, Fetcher (Google API) dan Create (Save ke DB)
	// TIDAK BOLEH dipanggil sama sekali. Kita buktikan di sini!
	mockFetcher.AssertNotCalled(t, "FetchByISBN", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestFetchAndSaveMetadata_BukuTidakAdaDiLokal_AmbilDariAPI(t *testing.T) {
	// 1. SIAPKAN STUNTMAN
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	bookUC := usecase.NewBookUsecase(mockRepo, mockFetcher)

	isbnTest := "9780134685991"
	fetchedBook := &domain.Book{
		ID:    "buku-baru-002",
		ISBN:  isbnTest,
		Title: "Effective Java",
	}

	// 2. ATUR SKENARIO (Ekspektasi)
	// Langkah A: Repo ditanya, tapi bukunya TIDAK ADA (return nil, nil)
	mockRepo.On("GetByISBN", mock.Anything, isbnTest).Return(nil, nil)

	// Langkah B: Karena tidak ada, Usecase PASTI memanggil Fetcher.
	// Kita suruh Fetcher pura-pura berhasil menemukan bukunya di Google.
	mockFetcher.On("FetchByISBN", mock.Anything, isbnTest).Return(fetchedBook, nil)

	// Langkah C: Karena berhasil di-fetch, Usecase PASTI memanggil Repo.Create untuk menyimpan.
	// Kita suruh Repo pura-pura berhasil menyimpan tanpa error.
	mockRepo.On("Create", mock.Anything, fetchedBook).Return(nil)

	// 3. EKSEKUSI ACTION!
	result, err := bookUC.FetchAndSaveMetadata(context.Background(), isbnTest)

	// 4. VALIDASI HASIL (Assert)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, fetchedBook.Title, result.Title)

	// 5. VALIDASI ALUR KERJA (Penting!)
	// Kita pastikan bahwa SEMUA skenario yang kita atur di atas (A, B, C)
	// BENAR-BENAR DIPANGGIL oleh usecase. Kalau ada satu saja yang terlewat, tes ini akan gagal.
	mockRepo.AssertExpectations(t)
	mockFetcher.AssertExpectations(t)
}

func TestFetchAndSaveMetadata_BukuTidakDitemukanDiManapun(t *testing.T) {
	// 1. KITA SIAPKAN STUNTMAN (PEMERAN PENGGANTI)
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	bookUC := usecase.NewBookUsecase(mockRepo, mockFetcher)

	isbnTest := "9999999999999" // ISBN ngasal yang pasti nggak ada

	// 2. KITA BERIKAN NASKAH SKENARIO KE STUNTMAN

	// Naskah A: Saat Otak nanya ke Repo Lokal, suruh Repo jawab "Nggak Ada (nil)"
	mockRepo.On("GetByISBN", mock.Anything, isbnTest).Return(nil, nil)

	// Naskah B: Saat Otak nanya ke Google API, suruh Google jawab "Nggak Ada (nil)" juga!
	mockFetcher.On("FetchByISBN", mock.Anything, isbnTest).Return(nil, nil)

	// 3. ACTION! (Mulai pengetesan)
	result, err := bookUC.FetchAndSaveMetadata(context.Background(), isbnTest)

	// 4. CEK HASIL AKHIRNYA
	assert.Error(t, err)  // Pastikan HARUS ADA error (karena gagal ketemu)
	assert.Nil(t, result) // Pastikan datanya kosong (nil)

	// Pastikan error-nya BENAR-BENAR error NotFound, bukan error database meledak
	assert.ErrorIs(t, err, domain.ErrNotFound)

	// 5. CEK KEDISIPLINAN OTAK (Sangat Penting!)
	// Karena buku nggak ketemu, Otak TIDAK BOLEH memanggil perintah Save/Create ke Database!
	// Kalau dia manggil Create, berarti Otaknya error/bug.
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// ==========================================
// TEST CREATE MANUAL (ADMIN)
// ==========================================

func TestCreateManual_ConflictISBN(t *testing.T) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher) // Tetap butuh meski tidak dipakai di fungsi ini
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	reqBook := &domain.Book{ISBN: "123", Title: "Buku Test"}

	// Naskah: ISBN ternyata sudah ada di DB
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

// ==========================================
// TEST UPDATE BOOK (ADMIN)
// ==========================================

func TestUpdateBook_NotFound(t *testing.T) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	// Naskah: Dicari by ID tidak ketemu
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

	// Naskah: Buku ketemu, lalu Update sukses
	mockRepo.On("GetByID", mock.Anything, "id-123").Return(oldBook, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Book")).Return(nil)

	res, err := uc.UpdateBook(context.Background(), "id-123", reqBook)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Baru", res.Title) // Pastikan title-nya terganti
}

// ==========================================
// TEST DELETE BOOK (ADMIN)
// ==========================================

func TestDeleteBook_NotFound(t *testing.T) {
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	// Naskah: Buku yang mau dihapus tidak ada
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

// ==========================================
// TEST GET ALL BOOKS (PAGINATION)
// ==========================================

func TestGetAllBooks_Sukses_MenghitungPaginasi(t *testing.T) {
	// 1. SIAPKAN STUNTMAN
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	// 2. SIAPKAN DATA PALSU DAN PARAMETER
	params := domain.PaginationQuery{
		Page:  1,
		Limit: 10,
	}

	mockBooks := []*domain.Book{
		{ID: "book-1", Title: "Buku Golang 101"},
		{ID: "book-2", Title: "Clean Architecture"},
	}

	// Skenario krusial: Total SELURUH buku di database ada 25
	var totalDataFromDB int64 = 25

	// 3. ATUR SKENARIO (Ekspektasi)
	// Kita perintahkan Stuntman Repo untuk mengembalikan 2 buku dan total data 25
	mockRepo.On("GetAll", mock.Anything, params).Return(mockBooks, totalDataFromDB, nil)

	// 4. EKSEKUSI ACTION!
	books, meta, err := uc.GetAllBooks(context.Background(), params)

	// 5. VALIDASI HASIL (Assert)
	assert.NoError(t, err)
	assert.NotNil(t, books)
	assert.Len(t, books, 2) // Pastikan data buku yang dikembalikan sesuai (2 buku)

	// 6. VALIDASI KALKULASI MATEMATIKA (Sangat Penting!)
	assert.Equal(t, int64(25), meta.TotalItems)
	assert.Equal(t, 10, meta.Limit)
	assert.Equal(t, 1, meta.CurrentPage)

	// Jika ada 25 buku, dan 1 halaman isinya 10,
	// maka total halamannya HARUS 3 (2.5 dibulatkan ke atas).
	assert.Equal(t, 3, meta.TotalPages)
}

func TestGetAllBooks_Gagal_DariRepository(t *testing.T) {
	// 1. SIAPKAN STUNTMAN
	mockRepo := new(mocks.BookRepository)
	mockFetcher := new(mocks.BookMetadataFetcher)
	uc := usecase.NewBookUsecase(mockRepo, mockFetcher)

	params := domain.PaginationQuery{Page: 1, Limit: 10}

	// Skenario: Database tiba-tiba mati / meledak
	dbError := domain.ErrInternalServerError // Atau error apapun dari repo

	// 2. ATUR SKENARIO
	mockRepo.On("GetAll", mock.Anything, params).Return(nil, int64(0), dbError)

	// 3. EKSEKUSI ACTION!
	books, meta, err := uc.GetAllBooks(context.Background(), params)

	// 4. VALIDASI HASIL
	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	assert.Nil(t, books)                // Data harus kosong
	assert.Equal(t, 0, meta.TotalPages) // Meta harus kosong karena gagal
}
