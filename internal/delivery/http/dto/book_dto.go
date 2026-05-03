package dto

type FetchBookRequest struct {
	ISBN string `json:"isbn" example:"9786020633176" validate:"required"`
}

type BookRequest struct {
	ISBN          string   `json:"isbn"`
	Title         string   `json:"title" validate:"required"`
	Authors       []string `json:"authors" validate:"required"`
	PublishedDate string   `json:"published_date"`
	Description   string   `json:"description"`
	PageCount     int      `json:"page_count"`
	CoverURL      string   `json:"cover_url"`
}
