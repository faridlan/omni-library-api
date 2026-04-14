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
	// 1. Konversi Domain ke Model DB
	dbModel := FromDomain(book)

	// 2. Simpan menggunakan dbModel, pastikan merujuk ke tabel "books"
	result := r.db.WithContext(ctx).Table("books").Create(dbModel)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *bookRepository) GetByISBN(ctx context.Context, isbn string) (*domain.Book, error) {
	var dbModel BookModel // Gunakan Model DB untuk menampung hasil query

	result := r.db.WithContext(ctx).Table("books").Where("isbn = ?", isbn).First(&dbModel)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	// 3. Konversi kembali ke Domain sebelum dikirim ke Usecase
	return dbModel.ToDomain(), nil
}
