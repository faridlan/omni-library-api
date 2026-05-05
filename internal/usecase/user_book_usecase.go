package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/faridlan/omni-library-api/internal/domain"
)

type userBookUsecase struct {
	userBookRepo domain.UserBookRepository
	bookRepo     domain.BookRepository
}

func NewUserBookUsecase(repo domain.UserBookRepository, bRepo domain.BookRepository) domain.UserBookUsecase {
	return &userBookUsecase{
		userBookRepo: repo,
		bookRepo:     bRepo,
	}
}

func (u *userBookUsecase) TrackNewBook(ctx context.Context, userID, bookID string) (*domain.UserBook, error) {

	_, err := u.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewError(domain.ErrNotFound, "Buku dengan ID tersebut tidak ditemukan")
		}
		return nil, err
	}

	// Gunakan inline assignment & abaikan ErrNotFound
	if existing, err := u.userBookRepo.FindByUserIDAndBookID(ctx, userID, bookID); err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err // Error internal DB
	} else if existing != nil {
		return nil, domain.NewError(domain.ErrConflict, "Buku sudah ada di rak pengguna")
	}

	newTrack := &domain.UserBook{
		UserID: userID,
		BookID: bookID,
		Status: "TO_READ",
	}

	err = u.userBookRepo.AddBookToShelf(ctx, newTrack)
	if err != nil {
		return nil, err
	}

	return newTrack, nil
}

func (u *userBookUsecase) UpdateReadingStatus(ctx context.Context, input domain.UpdateUserBookInput) (*domain.UserBook, error) {

	track, err := u.userBookRepo.FindByUserIDAndBookID(ctx, input.UserID, input.BookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewError(domain.ErrNotFound, "Buku dengan ID tersebut tidak ditemukan")
		}
		return nil, err
	}

	if input.Status != "" {
		track.Status = input.Status
	}
	if input.Page > 0 {
		track.CurrentPage = input.Page
	}
	if input.Rating >= 1 && input.Rating <= 5 {
		track.Rating = input.Rating
	}

	err = u.userBookRepo.UpdateProgress(ctx, &track.UserBook)
	if err != nil {
		return nil, err
	}

	return &track.UserBook, nil
}

func (u *userBookUsecase) GetUserLibrary(ctx context.Context, userID string, status string, params domain.PaginationQuery) ([]*domain.UserBookWithMetadata, domain.PaginationMeta, error) {
	books, totalItems, err := u.userBookRepo.FindAllByUserID(ctx, userID, status, params)
	if err != nil {
		return nil, domain.PaginationMeta{}, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(params.Limit)))

	meta := domain.PaginationMeta{
		CurrentPage: params.Page,
		Limit:       params.Limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
	}

	return books, meta, nil
}

func (u *userBookUsecase) GetUserBookDetail(ctx context.Context, userID, bookID string) (*domain.UserBookWithMetadata, error) {
	book, err := u.userBookRepo.GetDetailByID(ctx, userID, bookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewError(domain.ErrNotFound, "Buku dengan ID tersebut tidak ditemukan")
		}
		return nil, err
	}

	return book, nil
}

func (u *userBookUsecase) DeleteBookFromShelf(ctx context.Context, userID, bookID string) error {
	_, err := u.userBookRepo.FindByUserIDAndBookID(ctx, userID, bookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.NewError(domain.ErrNotFound, "Buku dengan ID tersebut tidak ditemukan")
		}
		return err
	}

	return u.userBookRepo.Delete(ctx, userID, bookID)
}
