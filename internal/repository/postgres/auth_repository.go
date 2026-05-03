package postgres

import (
	"context"
	"errors"

	"github.com/faridlan/omni-library-api/internal/domain"
	"gorm.io/gorm"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) domain.AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel

	result := r.db.WithContext(ctx).Where("email = ?", email).First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return model.ToDomain(), nil
}

func (r *authRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var model UserModel

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

func (r *authRepository) SaveRefreshToken(ctx context.Context, rt *domain.RefreshToken) error {
	model := RefreshTokenModel{
		UserID:    rt.UserID,
		Token:     rt.Token,
		ExpiresAt: rt.ExpiresAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *authRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var model RefreshTokenModel
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.RefreshToken{
		ID:        model.ID,
		UserID:    model.UserID,
		Token:     model.Token,
		ExpiresAt: model.ExpiresAt,
		CreatedAt: model.CreatedAt,
	}, nil
}

func (r *authRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&RefreshTokenModel{}).Error
}
