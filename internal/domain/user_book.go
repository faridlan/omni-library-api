package domain

import (
	"context"
	"time"
)

// ==========================================
// 1. ENTITY
// ==========================================
type UserBook struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	BookID      string    `json:"book_id"`
	Status      string    `json:"status"` // 'TO_READ', 'READING', 'FINISHED'
	CurrentPage int       `json:"current_page"`
	Rating      int       `json:"rating"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ==========================================
// 2. INTERFACES
// ==========================================
type UserBookRepository interface {
	// Menambahkan buku ke rak pribadi
	AddBookToShelf(ctx context.Context, ub *UserBook) error

	// Mengupdate progress bacaan (halaman atau status)
	UpdateProgress(ctx context.Context, ub *UserBook) error

	// Mengecek apakah buku sudah ada di rak user
	GetByUserAndBookID(ctx context.Context, userID, bookID string) (*UserBook, error)
}

type UserBookUsecase interface {
	// Fitur: User ingin memasukkan buku ke raknya (default status: TO_READ)
	TrackNewBook(ctx context.Context, userID, bookID string) (*UserBook, error)

	// Fitur: User ingin mengupdate dia sampai halaman berapa / kasih rating
	UpdateReadingStatus(ctx context.Context, userID, bookID, status string, page, rating int) (*UserBook, error)
}
