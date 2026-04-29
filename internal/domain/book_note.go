package domain

import (
	"context"
	"time"
)

// ==========================================
// 1. ENTITY
// ==========================================
type BookNote struct {
	ID            string    `json:"id"`
	UserBookID    string    `json:"user_book_id"` // Relasi ke buku di rak user
	Quote         string    `json:"quote"`
	PageReference int       `json:"page_reference"`
	Tags          []string  `json:"tags"` // Array murni Golang
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ==========================================
// 2. INTERFACES
// ==========================================
type BookNoteRepository interface {
	Create(ctx context.Context, note *BookNote) error
	// Mengambil semua catatan dari satu buku tertentu di rak
	GetByUserBookID(ctx context.Context, userBookID string, params PaginationQuery) ([]*BookNote, int64, error)
	GetByID(ctx context.Context, noteID string) (*BookNote, error)
	Delete(ctx context.Context, noteID string) error
	Update(ctx context.Context, note *BookNote) error
}

type BookNoteUsecase interface {
	AddNote(ctx context.Context, note *BookNote) error
	GetNotesForBook(ctx context.Context, userBookID string, params PaginationQuery) ([]*BookNote, PaginationMeta, error)
	DeleteNote(ctx context.Context, noteID string) error
	UpdateNote(ctx context.Context, note *BookNote) (*BookNote, error)
}
