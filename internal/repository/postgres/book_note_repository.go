package postgres

import (
	"context"

	"github.com/faridlan/omni-library-api/internal/domain"
	"gorm.io/gorm"
)

// Implementasi Repository
type bookNoteRepository struct {
	db *gorm.DB
}

func NewBookNoteRepository(db *gorm.DB) domain.BookNoteRepository {
	return &bookNoteRepository{db: db}
}

func (r *bookNoteRepository) Create(ctx context.Context, note *domain.BookNote) error {
	model := NoteFromDomain(note)

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}

	note.ID = model.ID
	return nil
}

func (r *bookNoteRepository) GetByUserBookID(ctx context.Context, userBookID string) ([]*domain.BookNote, error) {
	var models []BookNoteModel

	result := r.db.WithContext(ctx).Where("user_book_id = ?", userBookID).Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	var notes []*domain.BookNote
	for _, m := range models {
		mCopy := m
		notes = append(notes, mCopy.ToDomain())
	}

	return notes, nil
}
