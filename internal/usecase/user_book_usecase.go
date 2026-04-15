package usecase

import (
	"context"
	"errors"

	"github.com/faridlan/omni-library-api/internal/domain"
)

type userBookUsecase struct {
	userBookRepo domain.UserBookRepository
}

func NewUserBookUsecase(repo domain.UserBookRepository) domain.UserBookUsecase {
	return &userBookUsecase{
		userBookRepo: repo,
	}
}

func (u *userBookUsecase) TrackNewBook(ctx context.Context, userID, bookID string) (*domain.UserBook, error) {
	// ATURAN 1: Cek apakah buku sudah ada di rak user ini
	existing, err := u.userBookRepo.GetByUserAndBookID(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("buku sudah ada di dalam rak bacaan kamu")
	}

	// ATURAN 2: Jika belum ada, masukkan sebagai buku baru dengan status default
	newTrack := &domain.UserBook{
		UserID: userID,
		BookID: bookID,
		Status: "TO_READ", // Status awal selalu "Akan Dibaca"
	}

	err = u.userBookRepo.AddBookToShelf(ctx, newTrack)
	if err != nil {
		return nil, err
	}

	return newTrack, nil
}

func (u *userBookUsecase) UpdateReadingStatus(ctx context.Context, userID, bookID, status string, page, rating int) (*domain.UserBook, error) {
	// ATURAN 1: Pastikan bukunya ada di rak dia
	track, err := u.userBookRepo.GetByUserAndBookID(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}
	if track == nil {
		return nil, errors.New("buku tidak ditemukan di rak kamu")
	}

	// ATURAN 2: Update data yang boleh diubah
	if status != "" {
		track.Status = status
	}
	if page > 0 {
		track.CurrentPage = page
	}
	if rating >= 1 && rating <= 5 {
		track.Rating = rating
	}

	err = u.userBookRepo.UpdateProgress(ctx, track)
	if err != nil {
		return nil, err
	}

	return track, nil
}
