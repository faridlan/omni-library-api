package postgres

import (
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/lib/pq"
)

// DAO Khusus Database
type BookNoteModel struct {
	ID            string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserBookID    string `gorm:"type:uuid;not null"`
	Quote         string `gorm:"not null"`
	PageReference int
	Tags          pq.StringArray `gorm:"type:text[]"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Cegah GORM menebak nama tabel jadi "book_note_models"
func (BookNoteModel) TableName() string {
	return "book_notes"
}

// Konversi bolak-balik
func (m *BookNoteModel) ToDomain() *domain.BookNote {
	return &domain.BookNote{
		ID:            m.ID,
		UserBookID:    m.UserBookID,
		Quote:         m.Quote,
		PageReference: m.PageReference,
		Tags:          m.Tags,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func NoteFromDomain(d *domain.BookNote) *BookNoteModel {
	return &BookNoteModel{
		ID:            d.ID,
		UserBookID:    d.UserBookID,
		Quote:         d.Quote,
		PageReference: d.PageReference,
		Tags:          pq.StringArray(d.Tags), // Mapping ke array Postgres
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}
