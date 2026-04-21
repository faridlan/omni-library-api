package usecase

import (
	"context"

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
		return nil, domain.ErrNotFound
	}

	// ATURAN BISNIS 3: Buku ditemukan di internet! Simpan ke database lokal kita
	err = u.bookRepo.Create(ctx, newBook)
	if err != nil {
		return nil, err // Gagal menyimpan ke database
	}

	// Berhasil! Kembalikan buku yang baru saja disimpan
	return newBook, nil
}

// GetAllBooks mengambil seluruh katalog buku dari database lokal
func (u *bookUsecase) GetAllBooks(ctx context.Context) ([]*domain.Book, error) {
	// Memanggil fungsi GetAll yang baru saja kita buat di Repository
	books, err := u.bookRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return books, nil
}

func (u *bookUsecase) CreateManual(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	// 1. Cek apakah ISBN sudah dipakai (Jika Admin mengisi ISBN)
	if book.ISBN != "" {
		existing, _ := u.bookRepo.GetByISBN(ctx, book.ISBN)
		if existing != nil {
			return nil, domain.ErrConflict
		}
	}

	// 2. Simpan ke database
	err := u.bookRepo.Create(ctx, book)
	if err != nil {
		return nil, err
	}

	return book, nil
}

// Update Buku
func (u *bookUsecase) UpdateBook(ctx context.Context, id string, req *domain.Book) (*domain.Book, error) {
	// 1. Pastikan bukunya ada di database
	existing, err := u.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}

	// 2. Timpa data lama dengan data baru
	existing.Title = req.Title
	existing.Authors = req.Authors
	existing.Description = req.Description
	existing.PageCount = req.PageCount
	existing.CoverURL = req.CoverURL
	if req.ISBN != "" {
		existing.ISBN = req.ISBN
	}
	// (Tambahkan field lain jika perlu)

	// 3. Simpan perubahan
	err = u.bookRepo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete Buku
func (u *bookUsecase) DeleteBook(ctx context.Context, id string) error {
	// 1. Pastikan bukunya ada
	existing, err := u.bookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}

	// 2. Eksekusi hapus
	return u.bookRepo.Delete(ctx, id)
}
