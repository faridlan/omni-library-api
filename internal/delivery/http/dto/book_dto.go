package dto

// ==========================================
// DTO UNTUK REQUEST
// ==========================================

// FetchBookRequest adalah bentuk JSON yang kita harapkan dari Postman/Swagger
type FetchBookRequest struct {
	ISBN string `json:"isbn" example:"9786020633176" validate:"required"`
}

// DTO untuk Manual Input dan Update
type BookRequest struct {
	ISBN          string   `json:"isbn"` // Tidak wajib (karena buku indie kadang ga punya ISBN)
	Title         string   `json:"title" validate:"required"`
	Authors       []string `json:"authors" validate:"required"`
	PublishedDate string   `json:"published_date"`
	Description   string   `json:"description"`
	PageCount     int      `json:"page_count"`
	CoverURL      string   `json:"cover_url"`
}
