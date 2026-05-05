package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/faridlan/omni-library-api/internal/domain"
)

type bookNoteUsecase struct {
	noteRepo     domain.BookNoteRepository
	userBookRepo domain.UserBookRepository
}

func NewBookNoteUsecase(repo domain.BookNoteRepository, ubRepo domain.UserBookRepository) domain.BookNoteUsecase {
	return &bookNoteUsecase{
		noteRepo:     repo,
		userBookRepo: ubRepo,
	}
}

func (u *bookNoteUsecase) AddNote(ctx context.Context, input domain.CreateBookNoteInput) (*domain.BookNote, error) {

	_, err := u.userBookRepo.FindByID(ctx, input.UserBookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewError(domain.ErrNotFound, "Rak buku tidak ditemukan")
		}
		return nil, err
	}

	newNote := &domain.BookNote{
		UserBookID:    input.UserBookID,
		Quote:         input.Quote,
		PageReference: input.PageReference,
		Tags:          input.Tags,
	}

	err = u.noteRepo.Create(ctx, newNote)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	return newNote, nil
}

func (u *bookNoteUsecase) GetNotesForBook(ctx context.Context, userBookID string, params domain.PaginationQuery) ([]*domain.BookNote, domain.PaginationMeta, error) {

	_, err := u.userBookRepo.FindByID(ctx, userBookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.PaginationMeta{}, domain.NewError(domain.ErrNotFound, "Rak buku tidak ditemukan")
		}
		return nil, domain.PaginationMeta{}, err
	}

	notes, totalItems, err := u.noteRepo.FindAllByUserBookID(ctx, userBookID, params)
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

	return notes, meta, nil
}

func (u *bookNoteUsecase) DeleteNote(ctx context.Context, noteID string) error {

	_, err := u.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.NewError(domain.ErrNotFound, "Catatan tidak ditemukan")
		}
		return err
	}

	return u.noteRepo.Delete(ctx, noteID)
}

func (u *bookNoteUsecase) UpdateNote(ctx context.Context, input domain.UpdateBookNoteInput) (*domain.BookNote, error) {

	existing, err := u.noteRepo.FindByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewError(domain.ErrNotFound, "Catatan tidak ditemukan")
		}
		return nil, err
	}

	existing.Quote = input.Quote
	existing.PageReference = input.PageReference
	existing.Tags = input.Tags

	err = u.noteRepo.Update(ctx, existing)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	return existing, nil
}
