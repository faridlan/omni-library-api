package dto

type AddBookRequest struct {
	BookID string `json:"book_id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required,uuid"`
}

type UpdateProgressRequest struct {
	Status      string `json:"status" example:"READING"`
	CurrentPage int    `json:"current_page" example:"125"`
	Rating      int    `json:"rating" example:"5"`
}
