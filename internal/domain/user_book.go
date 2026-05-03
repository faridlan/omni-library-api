package domain

import (
	"context"
	"time"
)

type UserBook struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	BookID      string    `json:"book_id"`
	Status      string    `json:"status"`
	CurrentPage int       `json:"current_page"`
	Rating      int       `json:"rating"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserBookWithMetadata struct {
	UserBook
	Book Book `json:"book"`
}

type UserBookRepository interface {
	AddBookToShelf(ctx context.Context, ub *UserBook) error
	UpdateProgress(ctx context.Context, ub *UserBook) error
	GetByUserAndBookID(ctx context.Context, userID, bookID string) (*UserBookWithMetadata, error)
	GetByUserID(ctx context.Context, userID string, status string, params PaginationQuery) ([]*UserBookWithMetadata, int64, error)
	GetByID(ctx context.Context, id string) (*UserBook, error)
	Delete(ctx context.Context, userID, bookID string) error
	GetByBookID(ctx context.Context, userID, bookID string) (*UserBookWithMetadata, error)
}

type UserBookUsecase interface {
	TrackNewBook(ctx context.Context, userID, bookID string) (*UserBook, error)
	UpdateReadingStatus(ctx context.Context, userID, bookID, status string, page, rating int) (*UserBook, error)
	GetUserLibrary(ctx context.Context, userID string, status string, params PaginationQuery) ([]*UserBookWithMetadata, PaginationMeta, error)
	GetUserBookDetail(ctx context.Context, userID, bookID string) (*UserBookWithMetadata, error)
	DeleteBookFromShelf(ctx context.Context, userID, bookID string) error
}
