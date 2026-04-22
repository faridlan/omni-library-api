package postgres

import (
	"context"

	"github.com/faridlan/omni-library-api/internal/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository adalah constructor pembuat tangan repository
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// 1. Ubah data murni (Domain) menjadi data siap-database (Model)
	model := FromUserDomain(user)

	// 2. Suruh GORM menyimpannya ke PostgreSQL
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}

	// 3. Kembalikan ID dan waktu pembuatan yang dihasilkan oleh database ke Domain
	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt

	return nil
}
