package usecase

import (
	"context"
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

	masterBook, err := u.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	if masterBook == nil {
		return nil, domain.ErrNotFound
	}

	existing, err := u.userBookRepo.GetByBookID(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrConflict
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

func (u *userBookUsecase) UpdateReadingStatus(ctx context.Context, userID, bookID, status string, page, rating int) (*domain.UserBook, error) {

	track, err := u.userBookRepo.GetByUserAndBookID(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}
	if track == nil {
		return nil, domain.ErrNotFound
	}

	if status != "" {
		track.Status = status
	}
	if page > 0 {
		track.CurrentPage = page
	}
	if rating >= 1 && rating <= 5 {
		track.Rating = rating
	}

	err = u.userBookRepo.UpdateProgress(ctx, &track.UserBook)
	if err != nil {
		return nil, err
	}

	return &track.UserBook, nil
}

func (u *userBookUsecase) GetUserLibrary(ctx context.Context, userID string, status string, params domain.PaginationQuery) ([]*domain.UserBookWithMetadata, domain.PaginationMeta, error) {
	books, totalItems, err := u.userBookRepo.GetByUserID(ctx, userID, status, params)
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
	book, err := u.userBookRepo.GetByUserAndBookID(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, domain.ErrNotFound
	}

	return book, nil
}

func (u *userBookUsecase) DeleteBookFromShelf(ctx context.Context, userID, bookID string) error {
	existing, err := u.userBookRepo.GetByUserAndBookID(ctx, userID, bookID)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}

	return u.userBookRepo.Delete(ctx, userID, bookID)
}
