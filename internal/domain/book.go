package domain

import (
	"context"
	"time"
)

type Book struct {
	ID            string
	ISBN          string
	Title         string
	Authors       []string
	PublishedDate time.Time
	Description   string
	PageCount     int
	CoverURL      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateBookInput struct {
	ISBN          string
	Title         string
	Authors       []string
	PublishedDate time.Time
	Description   string
	PageCount     int
	CoverURL      string
}

type UpdateBookInput struct {
	ID            string
	ISBN          string
	Title         string
	Authors       []string
	PublishedDate time.Time
	Description   string
	PageCount     int
	CoverURL      string
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
	FetchAndSaveMetadata(ctx context.Context, isbn string) (*Book, error)
	GetAllBooks(ctx context.Context, params PaginationQuery) ([]*Book, PaginationMeta, error)
	GetBookByID(ctx context.Context, id string) (*Book, error)
	CreateManual(ctx context.Context, input CreateBookInput) (*Book, error)
	UpdateBook(ctx context.Context, input UpdateBookInput) (*Book, error)
	DeleteBook(ctx context.Context, id string) error
}

type BookMetadataFetcher interface {
	FetchByISBN(ctx context.Context, isbn string) (*Book, error)
}
