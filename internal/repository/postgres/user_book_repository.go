package postgres

import (
	"context"
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
	"gorm.io/gorm"
)

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

	ub.ID = model.ID
	return nil
}

func (r *userBookRepository) GetDetailByID(ctx context.Context, userID, userBookID string) (*domain.UserBookWithMetadata, error) {
	var model UserBookModel
	baseQuery := r.db.WithContext(ctx).Table("user_books").
		Where("user_id = ? AND id = ?", userID, userBookID)

	err := baseQuery.
		Preload("Book").
		Order("created_at DESC").
		First(&model).Error

	if err != nil {
		return nil, TranslateError(err)
	}

	bookResult := &domain.UserBookWithMetadata{
		UserBook: *model.ToDomain(),
		Book:     *model.Book.ToDomain(),
	}

	return bookResult, nil
}

func (r *userBookRepository) FindAllByUserID(ctx context.Context, userID string, status string, params domain.PaginationQuery) ([]*domain.UserBookWithMetadata, int64, error) {
	var dbModels []UserBookModel
	var totalItems int64

	baseQuery := r.db.WithContext(ctx).Model(&UserBookModel{}).Where("user_id = ?", userID)

	if status != "" {
		baseQuery = baseQuery.Where("status = ?", status)
	}

	if err := baseQuery.Count(&totalItems).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.
		Preload("Book").
		Limit(params.Limit).
		Offset(params.GetOffset()).
		Order("created_at DESC").
		Find(&dbModels).Error

	if err != nil {
		return nil, 0, err
	}

	var results []*domain.UserBookWithMetadata
	for _, m := range dbModels {
		results = append(results, &domain.UserBookWithMetadata{
			UserBook: *m.ToDomain(),
			Book:     *m.Book.ToDomain(),
		})
	}

	return results, totalItems, nil
}

func (r *userBookRepository) FindByID(ctx context.Context, id string) (*domain.UserBook, error) {
	var model UserBookModel
	err := r.db.WithContext(ctx).Table("user_books").Where("id = ?", id).First(&model).Error

	if err != nil {
		return nil, TranslateError(err)
	}

	return model.ToDomain(), nil
}

func (r *userBookRepository) FindByUserIDAndBookID(ctx context.Context, userID, bookID string) (*domain.UserBookWithMetadata, error) {
	var model UserBookModel
	err := r.db.WithContext(ctx).Table("user_books").
		Where("user_id = ? AND book_id = ?", userID, bookID).
		Preload("Book").
		First(&model).Error

	if err != nil {
		return nil, TranslateError(err)
	}

	bookResult := &domain.UserBookWithMetadata{
		UserBook: *model.ToDomain(),
		Book:     *model.Book.ToDomain(),
	}
	return bookResult, nil
}

func (r *userBookRepository) UpdateProgress(ctx context.Context, ub *domain.UserBook) error {

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

func (r *userBookRepository) Delete(ctx context.Context, userID, bookID string) error {
	result := r.db.WithContext(ctx).Table("user_books").Where("user_id = ? AND id = ?", userID, bookID).Delete(&UserBookModel{})
	return result.Error
}
