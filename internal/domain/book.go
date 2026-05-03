package domain

import (
	"context"
	"time"
)

type Book struct {
	ID            string    `json:"id"`
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

type BookRepository interface {
	GetByISBN(ctx context.Context, isbn string) (*Book, error)
	Create(ctx context.Context, book *Book) error
	GetAll(ctx context.Context, params PaginationQuery) ([]*Book, int64, error)
	GetByID(ctx context.Context, id string) (*Book, error)
	Update(ctx context.Context, book *Book) error
	Delete(ctx context.Context, id string) error
}

type BookUsecase interface {
	// Fitur 3: Mengambil data dari API luar dan menyimpannya
	FetchAndSaveMetadata(ctx context.Context, isbn string) (*Book, error)
	GetAllBooks(ctx context.Context, params PaginationQuery) ([]*Book, PaginationMeta, error)
	GetBookByID(ctx context.Context, id string) (*Book, error)
	CreateManual(ctx context.Context, book *Book) (*Book, error)
	UpdateBook(ctx context.Context, id string, req *Book) (*Book, error)
	DeleteBook(ctx context.Context, id string) error
}

type BookMetadataFetcher interface {
	FetchByISBN(ctx context.Context, isbn string) (*Book, error)
}
