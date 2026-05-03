package usecase

import (
	"context"
	"math"

	"github.com/faridlan/omni-library-api/internal/domain"
)

type bookUsecase struct {
	bookRepo domain.BookRepository
	fetcher  domain.BookMetadataFetcher
}

func NewBookUsecase(repo domain.BookRepository, fetcher domain.BookMetadataFetcher) domain.BookUsecase {
	return &bookUsecase{
		bookRepo: repo,
		fetcher:  fetcher,
	}
}

func (u *bookUsecase) FetchAndSaveMetadata(ctx context.Context, isbn string) (*domain.Book, error) {

	existingBook, err := u.bookRepo.GetByISBN(ctx, isbn)
	if err != nil {
		return nil, err
	}

	if existingBook != nil {
		return existingBook, nil
	}

	newBook, err := u.fetcher.FetchByISBN(ctx, isbn)
	if err != nil {
		return nil, err
	}

	if newBook == nil {
		return nil, domain.ErrNotFound
	}

	err = u.bookRepo.Create(ctx, newBook)
	if err != nil {
		return nil, err
	}

	return newBook, nil
}

func (u *bookUsecase) GetAllBooks(ctx context.Context, params domain.PaginationQuery) ([]*domain.Book, domain.PaginationMeta, error) {

	books, totalItems, err := u.bookRepo.GetAll(ctx, params)
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

func (u *bookUsecase) GetBookByID(ctx context.Context, id string) (*domain.Book, error) {
	book, err := u.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, domain.ErrNotFound
	}
	return book, nil
}

func (u *bookUsecase) CreateManual(ctx context.Context, book *domain.Book) (*domain.Book, error) {

	if book.ISBN != "" {
		existing, _ := u.bookRepo.GetByISBN(ctx, book.ISBN)
		if existing != nil {
			return nil, domain.ErrConflict
		}
	}

	err := u.bookRepo.Create(ctx, book)
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (u *bookUsecase) UpdateBook(ctx context.Context, id string, req *domain.Book) (*domain.Book, error) {

	existing, err := u.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}

	existing.Title = req.Title
	existing.Authors = req.Authors
	existing.Description = req.Description
	existing.PageCount = req.PageCount
	existing.CoverURL = req.CoverURL
	if req.ISBN != "" {
		existing.ISBN = req.ISBN
	}

	err = u.bookRepo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (u *bookUsecase) DeleteBook(ctx context.Context, id string) error {
	existing, err := u.bookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}

	return u.bookRepo.Delete(ctx, id)
}
