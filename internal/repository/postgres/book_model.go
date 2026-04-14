package postgres

import (
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/lib/pq"
)

// BookModel adalah representasi murni dari tabel 'books' di PostgreSQL.
// Kita menggunakan pq.StringArray agar GORM bisa menyimpannya sebagai TEXT[].
type BookModel struct {
	ID            string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ISBN          string `gorm:"unique"`
	Title         string
	Authors       pq.StringArray `gorm:"type:text[]"`
	PublishedDate time.Time
	Description   string
	PageCount     int
	CoverURL      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ToDomain mengonversi dari Model Database kembali menjadi Model Domain murni
func (m *BookModel) ToDomain() *domain.Book {
	return &domain.Book{
		ID:            m.ID,
		ISBN:          m.ISBN,
		Title:         m.Title,
		Authors:       m.Authors, // pq.StringArray pada dasarnya adalah []string
		PublishedDate: m.PublishedDate,
		Description:   m.Description,
		PageCount:     m.PageCount,
		CoverURL:      m.CoverURL,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// FromDomain mengonversi dari Model Domain menjadi Model Database
func FromDomain(d *domain.Book) *BookModel {
	return &BookModel{
		ID:            d.ID,
		ISBN:          d.ISBN,
		Title:         d.Title,
		Authors:       pq.StringArray(d.Authors), // Konversi []string Golang ke tipe array Postgres
		PublishedDate: d.PublishedDate,
		Description:   d.Description,
		PageCount:     d.PageCount,
		CoverURL:      d.CoverURL,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}
