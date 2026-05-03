package usecase

import (
	"context"
	"errors"

	"github.com/faridlan/omni-library-api/internal/delivery/http/dto"
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(userRepo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (u *userUsecase) UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) (*domain.User, error) {
	if err := utils.ValidateStruct(req); err != nil {
		return nil, err
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	user.Name = req.Name

	if err := u.userRepo.Update(ctx, user); err != nil {
		// return nil, errors.New("gagal memperbarui profil")
		return nil, domain.ErrBadParamInput
	}

	return user, nil
}

func (u *userUsecase) UpdatePassword(ctx context.Context, userID string, req *dto.UpdatePasswordRequest) error {

	if err := utils.ValidateStruct(req); err != nil {
		return domain.ErrBadParamInput
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return domain.ErrNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		return domain.ErrBadParamInput
		// return errors.New("password lama salah")
	}

	if req.OldPassword == req.NewPassword {
		return domain.ErrBadParamInput
		// return errors.New("password baru tidak boleh sama dengan password lama")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrBadParamInput
		// return errors.New("gagal memproses password baru")
	}

	user.Password = string(hashedPassword)
	if err := u.userRepo.Update(ctx, user); err != nil {
		return errors.New("gagal memperbarui password")
	}

	return nil
}
