package usecase

import (
	"context"
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

func (u *bookNoteUsecase) AddNote(ctx context.Context, note *domain.BookNote) error {

	userBook, err := u.userBookRepo.GetByID(ctx, note.UserBookID)
	if err != nil {
		return err
	}

	if userBook == nil {
		return domain.ErrNotFound
	}

	// Lanjut simpan ke database
	return u.noteRepo.Create(ctx, note)
}

func (u *bookNoteUsecase) GetNotesForBook(ctx context.Context, userBookID string, params domain.PaginationQuery) ([]*domain.BookNote, domain.PaginationMeta, error) {
	// Ambil semua catatan berdasarkan ID buku di rak user

	userBook, err := u.userBookRepo.GetByID(ctx, userBookID)
	if err != nil {
		return nil, domain.PaginationMeta{}, err
	}
	if userBook == nil {
		return nil, domain.PaginationMeta{}, domain.ErrNotFound // Berhenti di sini!
	}

	notes, totalItems, err := u.noteRepo.GetByUserBookID(ctx, userBookID, params)
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
