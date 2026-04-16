package usecase

import (
	"context"

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

func (u *bookNoteUsecase) GetNotesForBook(ctx context.Context, userBookID string) ([]*domain.BookNote, error) {
	// Ambil semua catatan berdasarkan ID buku di rak user
	return u.noteRepo.GetByUserBookID(ctx, userBookID)
}
