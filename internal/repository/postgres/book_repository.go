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

	book.ID = dbModel.ID
	book.CreatedAt = dbModel.CreatedAt
	book.UpdatedAt = dbModel.UpdatedAt

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

// GetAll mengambil seluruh data buku dari tabel 'books'
func (r *bookRepository) GetAll(ctx context.Context) ([]*domain.Book, error) {
	var dbModels []BookModel // Menampung hasil dari GORM/Postgres

	// Gunakan Find() untuk mengambil semua baris data
	result := r.db.WithContext(ctx).Table("books").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	// Kita buat slice kosong untuk menampung data Domain murni
	var books []*domain.Book

	// Looping untuk mengonversi setiap dbModel kembali menjadi Domain murni
	for _, model := range dbModels {
		// Karena range di Golang menggunakan pass-by-value, kita butuh variable lokal
		m := model
		books = append(books, m.ToDomain())
	}

	return books, nil
}

func (r *bookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	var model BookModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Buku tidak ditemukan, kembalikan nil
		}
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

// Update data buku
func (r *bookRepository) Update(ctx context.Context, book *domain.Book) error {
	model := FromDomain(book)
	// Gunakan Updates() agar GORM hanya mengupdate field yang tidak kosong/berubah
	return r.db.WithContext(ctx).Model(&BookModel{}).Where("id = ?", model.ID).Updates(model).Error
}

// Hapus buku secara permanen (Hard Delete)
// Catatan: Karena di SQL kamu pasang ON DELETE CASCADE, menghapus buku ini
// akan otomatis menghapus rak (user_books) dan catatan (book_notes) yang terkait!
func (r *bookRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&BookModel{}).Error
}
