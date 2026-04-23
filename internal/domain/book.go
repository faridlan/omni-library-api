package domain

import (
	"context"
	"time"
)

// ==========================================
// 1. ENTITY (Representasi Data)
// ==========================================

// Book merepresentasikan tabel books di database kita.
// Kita pakai tag JSON agar nanti otomatis rapi saat dikirim sebagai response API.
type Book struct {
	ID            string    `json:"id"` // Menggunakan string untuk menampung UUID
	ISBN          string    `json:"isbn"`
	Title         string    `json:"title"`
	Authors       []string  `json:"authors"`
	PublishedDate time.Time `json:"published_date"`
	Description   string    `json:"description"`
	PageCount     int       `json:"page_count"`
	CoverURL      string    `json:"cover_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ==========================================
// 2. INTERFACE (Kontrak Kerja)
// ==========================================

// BookRepository adalah kontrak untuk layer Repository (yang ngobrol ke Postgres).
// Siapapun yang menjadi repository buku, WAJIB punya fungsi-fungsi ini.
type BookRepository interface {
	GetByISBN(ctx context.Context, isbn string) (*Book, error)
	Create(ctx context.Context, book *Book) error
	GetAll(ctx context.Context, params PaginationQuery) ([]*Book, int64, error)
	GetByID(ctx context.Context, id string) (*Book, error)
	Update(ctx context.Context, book *Book) error
	Delete(ctx context.Context, id string) error
}

// BookUsecase adalah kontrak untuk layer Usecase (otak bisnis kita).
// Di sini kita definisikan fitur utama yang akan kita buat malam ini.
type BookUsecase interface {
	// Fitur 3: Mengambil data dari API luar dan menyimpannya
	FetchAndSaveMetadata(ctx context.Context, isbn string) (*Book, error)
	GetAllBooks(ctx context.Context, params PaginationQuery) ([]*Book, PaginationMeta, error)
	CreateManual(ctx context.Context, book *Book) (*Book, error)
	UpdateBook(ctx context.Context, id string, req *Book) (*Book, error)
	DeleteBook(ctx context.Context, id string) error
}

// BookMetadataFetcher adalah kontrak untuk mengambil data dari API eksternal (Google Books)
type BookMetadataFetcher interface {
	FetchByISBN(ctx context.Context, isbn string) (*Book, error)
}
