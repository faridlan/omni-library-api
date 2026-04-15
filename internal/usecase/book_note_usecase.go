package usecase

import (
	"context"
	"errors"

	"github.com/faridlan/omni-library-api/internal/domain"
)

type bookNoteUsecase struct {
	noteRepo domain.BookNoteRepository
}

func NewBookNoteUsecase(repo domain.BookNoteRepository) domain.BookNoteUsecase {
	return &bookNoteUsecase{
		noteRepo: repo,
	}
}

func (u *bookNoteUsecase) AddNote(ctx context.Context, note *domain.BookNote) error {
	// ATURAN BISNIS: Quote wajib diisi
	if note.Quote == "" {
		return errors.New("kutipan (quote) tidak boleh kosong")
	}

	// Lanjut simpan ke database
	return u.noteRepo.Create(ctx, note)
}

func (u *bookNoteUsecase) GetNotesForBook(ctx context.Context, userBookID string) ([]*domain.BookNote, error) {
	// Ambil semua catatan berdasarkan ID buku di rak user
	return u.noteRepo.GetByUserBookID(ctx, userBookID)
}
