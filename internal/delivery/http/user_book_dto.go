package http

// ==========================================
// DTO UNTUK REQUEST LIBRARY
// ==========================================

// AddBookRequest adalah payload untuk menambahkan buku ke rak
type AddBookRequest struct {
	BookID string `json:"book_id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
}

// UpdateProgressRequest adalah payload untuk mengupdate progres bacaan
type UpdateProgressRequest struct {
	Status      string `json:"status" example:"READING"`
	CurrentPage int    `json:"current_page" example:"125"`
	Rating      int    `json:"rating" example:"5"`
}
