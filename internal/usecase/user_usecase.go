package usecase

import (
	"context"
	"errors"

	"github.com/faridlan/omni-library-api/internal/domain"
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
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewError(domain.ErrNotFound, "User dengan ID tersebut tidak ditemukan")
		}
		return nil, err
	}
	return user, nil
}

func (u *userUsecase) UpdateProfile(ctx context.Context, input domain.UpdateProfileInput) (*domain.User, error) {
	user, err := u.userRepo.FindByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewError(domain.ErrNotFound, "User dengan ID tersebut tidak ditemukan")
		}
		return nil, err
	}

	user.Name = input.Name

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) UpdatePassword(ctx context.Context, input domain.UpdatePasswordInput) error {

	user, err := u.userRepo.FindByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.NewError(domain.ErrNotFound, "User dengan ID tersebut tidak ditemukan")
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword))
	if err != nil {
		return domain.NewError(domain.ErrBadParamInput, "Password lama salah")
	}

	if input.OldPassword == input.NewPassword {
		return domain.NewError(domain.ErrBadParamInput, "Password baru tidak boleh sama dengan password lama")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrInternalServerError
	}

	user.Password = string(hashedPassword)
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}
