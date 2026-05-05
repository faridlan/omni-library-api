package usecase

import (
	"context"
	"errors"
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

	if existingBook, err := u.bookRepo.GetByISBN(ctx, isbn); err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	} else if existingBook != nil {
		return nil, domain.NewError(domain.ErrConflict, "Buku dengan ISBN tersebut sudah ada di sistem")
	}

	newBook, err := u.fetcher.FetchByISBN(ctx, isbn)
	if err != nil {
		return nil, err
	}

	if newBook == nil {
		return nil, domain.NewError(domain.ErrNotFound, "Buku dengan ISBN tersebut tidak ditemukan")
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
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewError(domain.ErrNotFound, "Buku dengan ID tersebut tidak ditemukan")
		}
		return nil, err
	}
	return book, nil
}

func (u *bookUsecase) CreateManual(ctx context.Context, input domain.CreateBookInput) (*domain.Book, error) {

	if input.ISBN != "" {
		if existing, err := u.bookRepo.GetByISBN(ctx, input.ISBN); err != nil && !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		} else if existing != nil {
			return nil, domain.NewError(domain.ErrConflict, "Buku dengan ISBN tersebut sudah terdaftar")
		}
	}

	bookInput := &domain.Book{
		ISBN:          input.ISBN,
		Title:         input.Title,
		Authors:       input.Authors,
		PublishedDate: input.PublishedDate,
		Description:   input.Description,
		PageCount:     input.PageCount,
		CoverURL:      input.CoverURL,
	}

	err := u.bookRepo.Create(ctx, bookInput)
	if err != nil {
		return nil, err
	}

	return bookInput, nil
}

func (u *bookUsecase) UpdateBook(ctx context.Context, input domain.UpdateBookInput) (*domain.Book, error) {

	existing, err := u.bookRepo.GetByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewError(domain.ErrNotFound, "Buku dengan ID tersebut tidak ditemukan")
		}
		return nil, err
	}

	existing.Title = input.Title
	existing.Authors = input.Authors
	existing.Description = input.Description
	existing.PageCount = input.PageCount
	existing.CoverURL = input.CoverURL
	if input.ISBN != "" {
		existing.ISBN = input.ISBN
	}

	err = u.bookRepo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (u *bookUsecase) DeleteBook(ctx context.Context, id string) error {
	_, err := u.bookRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.NewError(domain.ErrNotFound, "Buku dengan ID tersebut tidak ditemukan")
		}
		return err
	}

	return u.bookRepo.Delete(ctx, id)
}
