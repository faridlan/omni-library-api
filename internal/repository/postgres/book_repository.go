package postgres

import (
	"context"

	"github.com/faridlan/omni-library-api/internal/domain"
	"gorm.io/gorm"
)

type bookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) domain.BookRepository {
	return &bookRepository{
		db: db,
	}
}

func (r *bookRepository) Create(ctx context.Context, book *domain.Book) error {
	dbModel := FromDomain(book)

	result := r.db.WithContext(ctx).Table("books").Create(dbModel)
	if result.Error != nil {
		return result.Error
	}

	book.ID = dbModel.ID
	book.CreatedAt = dbModel.CreatedAt
	book.UpdatedAt = dbModel.UpdatedAt

	return nil
}

func (r *bookRepository) GetByISBN(ctx context.Context, isbn string) (*domain.Book, error) {
	var dbModel BookModel

	err := r.db.WithContext(ctx).Table("books").Where("isbn = ?", isbn).First(&dbModel).Error

	if err != nil {
		return nil, TranslateError(err)
	}

	return dbModel.ToDomain(), nil
}

func (r *bookRepository) GetAll(ctx context.Context, params domain.PaginationQuery) ([]*domain.Book, int64, error) {
	var dbModels []BookModel
	var totalItems int64

	err := r.db.WithContext(ctx).Model(&BookModel{}).Count(&totalItems).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Limit(params.Limit).
		Offset(params.GetOffset()).
		Order("created_at DESC").
		Find(&dbModels).Error

	if err != nil {
		return nil, 0, err
	}

	var books []*domain.Book
	for _, model := range dbModels {
		m := model
		books = append(books, m.ToDomain())
	}

	return books, totalItems, nil
}

func (r *bookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	var model BookModel

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error

	if err != nil {
		return nil, TranslateError(err)
	}

	return model.ToDomain(), nil
}

func (r *bookRepository) Update(ctx context.Context, book *domain.Book) error {
	model := FromDomain(book)

	return r.db.WithContext(ctx).Model(&BookModel{}).Where("id = ?", model.ID).Updates(model).Error
}

func (r *bookRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&BookModel{}).Error
}
