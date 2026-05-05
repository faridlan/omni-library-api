package postgres

import (
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
)

type UserBookModel struct {
	ID          string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID      string `gorm:"type:uuid;not null"`
	BookID      string `gorm:"type:uuid;not null"`
	Status      string
	CurrentPage int
	Rating      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Book        BookModel `gorm:"foreignKey:BookID"`
}

func (UserBookModel) TableName() string {
	return "user_books"
}

func (m *UserBookModel) ToDomain() *domain.UserBook {
	return &domain.UserBook{
		ID:          m.ID,
		UserID:      m.UserID,
		BookID:      m.BookID,
		Status:      m.Status,
		CurrentPage: m.CurrentPage,
		Rating:      m.Rating,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
