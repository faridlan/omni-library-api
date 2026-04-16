package postgres

import (
	"context"
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
	"gorm.io/gorm"
)

// Implementasi Repository
type userBookRepository struct {
	db *gorm.DB
}

func NewUserBookRepository(db *gorm.DB) domain.UserBookRepository {
	return &userBookRepository{db: db}
}

func (r *userBookRepository) AddBookToShelf(ctx context.Context, ub *domain.UserBook) error {
	model := &UserBookModel{
		UserID: ub.UserID,
		BookID: ub.BookID,
		Status: ub.Status,
	}
	result := r.db.WithContext(ctx).Table("user_books").Omit("CurrentPage", "Rating").Create(model)
	if result.Error != nil {
		return result.Error
	}
	// Kembalikan ID yang di-generate Postgres ke domain
	ub.ID = model.ID
	return nil
}

func (r *userBookRepository) GetByUserAndBookID(ctx context.Context, userID, bookID string) (*domain.UserBook, error) {
	var model UserBookModel
	result := r.db.WithContext(ctx).Table("user_books").
		Where("user_id = ? AND book_id = ?", userID, bookID).First(&model)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return model.ToDomain(), nil
}

func (r *userBookRepository) UpdateProgress(ctx context.Context, ub *domain.UserBook) error {
	// Update hanya field yang relevan berdasarkan ID
	updateData := map[string]any{
		"status":       ub.Status,
		"current_page": ub.CurrentPage,
		"updated_at":   time.Now(),
	}

	if ub.Rating >= 1 && ub.Rating <= 5 {
		updateData["rating"] = ub.Rating
	}

	result := r.db.WithContext(ctx).Table("user_books").
		Where("id = ?", ub.ID).
		Updates(updateData)

	return result.Error
}

func (r *userBookRepository) GetByUserID(ctx context.Context, userID string, status string) ([]*domain.UserBookWithMetadata, error) {
	var dbModels []UserBookModel

	// Kita gunakan Preload("Book") agar GORM otomatis melakukan JOIN ke tabel books
	query := r.db.WithContext(ctx).Table("user_books").
		Preload("Book").
		Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	result := query.Find(&dbModels)

	if result.Error != nil {
		return nil, result.Error
	}

	var results []*domain.UserBookWithMetadata
	for _, m := range dbModels {
		results = append(results, &domain.UserBookWithMetadata{
			UserBook: *m.ToDomain(),
			Book:     *m.Book.ToDomain(), // Konversi BookModel ke Domain Book
		})
	}

	return results, nil
}

func (r *userBookRepository) GetByID(ctx context.Context, id string) (*domain.UserBook, error) {
	var model UserBookModel
	result := r.db.WithContext(ctx).Table("user_books").Where("id = ?", id).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return model.ToDomain(), nil
}
