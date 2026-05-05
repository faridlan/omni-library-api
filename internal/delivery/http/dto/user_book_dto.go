package dto

import "time"

type AddBookRequest struct {
	BookID string `json:"book_id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required,uuid"`
}

type UpdateProgressRequest struct {
	Status      string `json:"status" example:"READING"`
	CurrentPage int    `json:"current_page" example:"125"`
	Rating      int    `json:"rating" example:"5"`
}

type UserBookResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	BookID      string    `json:"book_id"`
	Status      string    `json:"status"`
	CurrentPage int       `json:"current_page"`
	Rating      int       `json:"rating"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserBookWithMetaDataResponse struct {
	UserBookResponse
	Book BookResponse
}
