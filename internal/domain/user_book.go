package domain

import (
	"context"
	"time"
)

type UserBook struct {
	ID          string
	UserID      string
	BookID      string
	Status      string
	CurrentPage int
	Rating      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserBookWithMetadata struct {
	UserBook
	Book Book `json:"book"`
}

type UpdateUserBookInput struct {
	ID     string
	UserID string
	BookID string
	Status string
	Page   int
	Rating int
}

type UserBookRepository interface {
	AddBookToShelf(ctx context.Context, ub *UserBook) error
	UpdateProgress(ctx context.Context, ub *UserBook) error
	GetDetailByID(ctx context.Context, userID, userBookID string) (*UserBookWithMetadata, error)
	FindAllByUserID(ctx context.Context, userID string, status string, params PaginationQuery) ([]*UserBookWithMetadata, int64, error)
	FindByID(ctx context.Context, id string) (*UserBook, error)
	FindByUserIDAndBookID(ctx context.Context, userID, bookID string) (*UserBookWithMetadata, error)
	Delete(ctx context.Context, userID, bookID string) error
}

type UserBookUsecase interface {
	TrackNewBook(ctx context.Context, userID, bookID string) (*UserBook, error)
	UpdateReadingStatus(ctx context.Context, input UpdateUserBookInput) (*UserBook, error)
	GetUserLibrary(ctx context.Context, userID string, status string, params PaginationQuery) ([]*UserBookWithMetadata, PaginationMeta, error)
	GetUserBookDetail(ctx context.Context, userID, bookID string) (*UserBookWithMetadata, error)
	DeleteBookFromShelf(ctx context.Context, userID, bookID string) error
}
