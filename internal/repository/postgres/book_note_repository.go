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

func (r *bookNoteRepository) GetByUserBookID(ctx context.Context, userBookID string, params domain.PaginationQuery) ([]*domain.BookNote, int64, error) {
	var models []BookNoteModel
	var totalItems int64

	// 1. Buat Base Query (Kondisi Utama)
	baseQuery := r.db.WithContext(ctx).Model(&BookNoteModel{}).Where("user_book_id = ?", userBookID)

	// 2. Hitung Total (Berdasarkan kondisi di atas)
	if err := baseQuery.Count(&totalItems).Error; err != nil {
		return nil, 0, err
	}

	// 3. Ambil Data (Kondisi WHERE sudah menempel di baseQuery)
	err := baseQuery.
		Limit(params.Limit).
		Offset(params.GetOffset()).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, 0, err
	}

	var notes []*domain.BookNote
	for _, m := range models {
		mCopy := m
		notes = append(notes, mCopy.ToDomain())
	}

	return notes, totalItems, nil
}
