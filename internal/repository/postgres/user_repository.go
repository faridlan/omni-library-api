package postgres

import (
	"context"
	"errors"

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

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel

	// Cari 1 baris pertama yang email-nya cocok
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&model)

	if result.Error != nil {
		// Jika error-nya karena "Data tidak ditemukan", jangan anggap ini sebagai sistem crash (500).
		// Kembalikan (nil, nil) agar Usecase tahu bahwa email ini belum terdaftar.
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Jika error lain (misal koneksi putus), lemparkan error-nya
		return nil, result.Error
	}

	// Jika ketemu, ubah kembali dari Model ke data murni (Domain)
	return model.ToDomain(), nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var model UserModel

	// Cari 1 baris pertama yang ID-nya cocok
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

// Simpan token baru saat Login
func (r *userRepository) SaveRefreshToken(ctx context.Context, rt *domain.RefreshToken) error {
	model := RefreshTokenModel{
		UserID:    rt.UserID,
		Token:     rt.Token,
		ExpiresAt: rt.ExpiresAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

// Cari token (Nanti dipakai saat endpoint /refresh)
func (r *userRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var model RefreshTokenModel
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Kembalikan nil jika token tidak ditemukan (Palsu/Sudah dihapus)
		}
		return nil, err
	}

	// Mapping balik ke Domain
	return &domain.RefreshToken{
		ID:        model.ID,
		UserID:    model.UserID,
		Token:     model.Token,
		ExpiresAt: model.ExpiresAt,
		CreatedAt: model.CreatedAt,
	}, nil
}

// Hapus token (Nanti dipakai saat Logout atau strategi Rotation)
func (r *userRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&RefreshTokenModel{}).Error
}
