package usecase

import (
	"context"
	"errors"

	"github.com/faridlan/omni-library-api/internal/domain"
)

// bookUsecase adalah implementasi dari domain.BookUsecase
type bookUsecase struct {
	bookRepo domain.BookRepository
	fetcher  domain.BookMetadataFetcher
}

// NewBookUsecase menginjeksi repository dan fetcher ke dalam usecase
func NewBookUsecase(repo domain.BookRepository, fetcher domain.BookMetadataFetcher) domain.BookUsecase {
	return &bookUsecase{
		bookRepo: repo,
		fetcher:  fetcher,
	}
}

// FetchAndSaveMetadata adalah fitur utama kita malam ini
func (u *bookUsecase) FetchAndSaveMetadata(ctx context.Context, isbn string) (*domain.Book, error) {
	// ATURAN BISNIS 1: Cek apakah buku sudah ada di database lokal kita?
	// Ini penerapan efisiensi agar kita tidak buang-buang kuota API eksternal
	existingBook, err := u.bookRepo.GetByISBN(ctx, isbn)
	if err != nil {
		return nil, err // Return error jika database sedang bermasalah
	}

	// Jika buku sudah ada, langsung kembalikan data dari database lokal
	if existingBook != nil {
		return existingBook, nil
	}

	// ATURAN BISNIS 2: Buku tidak ada di lokal. Saatnya minta tolong fetcher cari di internet
	newBook, err := u.fetcher.FetchByISBN(ctx, isbn)
	if err != nil {
		return nil, err // Error saat koneksi ke Google Books
	}

	// Jika Google Books tidak punya bukunya
	if newBook == nil {
		return nil, errors.New("buku tidak ditemukan di database publik")
	}

	// ATURAN BISNIS 3: Buku ditemukan di internet! Simpan ke database lokal kita
	err = u.bookRepo.Create(ctx, newBook)
	if err != nil {
		return nil, err // Gagal menyimpan ke database
	}

	// Berhasil! Kembalikan buku yang baru saja disimpan
	return newBook, nil
}
