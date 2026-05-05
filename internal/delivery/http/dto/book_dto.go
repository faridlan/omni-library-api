package dto

import "time"

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

type BookResponse struct {
	ID            string    `json:"id"`
	ISBN          string    `json:"isbn"`
	Title         string    `json:"title"`
	Authors       []string  `json:"authors"`
	PublishedDate time.Time `json:"published_date"`
	Description   string    `json:"description"`
	PageCount     int       `json:"page_count"`
	CoverURL      string    `json:"cover_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
