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
