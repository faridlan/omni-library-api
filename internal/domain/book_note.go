package domain

import (
	"context"
	"time"
)

type BookNote struct {
	ID            string
	UserBookID    string
	Quote         string
	PageReference int
	Tags          []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateBookNoteInput struct {
	UserBookID    string
	Quote         string
	PageReference int
	Tags          []string
}

type UpdateBookNoteInput struct {
	ID            string
	Quote         string
	PageReference int
	Tags          []string
}

type BookNoteRepository interface {
	Create(ctx context.Context, note *BookNote) error
	FindAllByUserBookID(ctx context.Context, userBookID string, params PaginationQuery) ([]*BookNote, int64, error)
	FindByID(ctx context.Context, noteID string) (*BookNote, error)
	Update(ctx context.Context, note *BookNote) error
	Delete(ctx context.Context, noteID string) error
}

type BookNoteUsecase interface {
	AddNote(ctx context.Context, input CreateBookNoteInput) (*BookNote, error)
	GetNotesForBook(ctx context.Context, userBookID string, params PaginationQuery) ([]*BookNote, PaginationMeta, error)
	UpdateNote(ctx context.Context, input UpdateBookNoteInput) (*BookNote, error)
	DeleteNote(ctx context.Context, noteID string) error
}
