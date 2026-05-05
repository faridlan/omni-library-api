package postgres

import (
	"context"

	"github.com/faridlan/omni-library-api/internal/domain"
	"gorm.io/gorm"
)

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
	note.CreatedAt = model.CreatedAt
	note.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *bookNoteRepository) FindAllByUserBookID(ctx context.Context, userBookID string, params domain.PaginationQuery) ([]*domain.BookNote, int64, error) {
	var models []BookNoteModel
	var totalItems int64

	baseQuery := r.db.WithContext(ctx).Model(&BookNoteModel{}).Where("user_book_id = ?", userBookID)

	if err := baseQuery.Count(&totalItems).Error; err != nil {
		return nil, 0, err
	}

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

func (r *bookNoteRepository) FindByID(ctx context.Context, noteID string) (*domain.BookNote, error) {
	var model BookNoteModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", noteID).Error

	if err != nil {
		return nil, TranslateError(err)
	}

	return model.ToDomain(), nil
}

func (r *bookNoteRepository) Update(ctx context.Context, note *domain.BookNote) error {
	model := NoteFromDomain(note)

	err := r.db.WithContext(ctx).Save(model).Error
	if err != nil {
		return err
	}

	note.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *bookNoteRepository) Delete(ctx context.Context, noteID string) error {
	result := r.db.WithContext(ctx).Delete(&BookNoteModel{}, "id = ?", noteID)
	return result.Error
}
