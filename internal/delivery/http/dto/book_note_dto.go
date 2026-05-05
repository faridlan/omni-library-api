package dto

import "time"

type AddNoteRequest struct {
	Quote         string   `json:"quote" example:"Bekerjalah seperti programmer pemalas..." validate:"required"`
	PageReference int      `json:"page_reference" example:"42"`
	Tags          []string `json:"tags" example:"Inspiratif,Programming"`
}

type UpdateNoteRequest struct {
	Quote         string   `json:"quote" example:"Bekerjalah seperti programmer pemalas..." validate:"required"`
	PageReference int      `json:"page_reference" example:"42"`
	Tags          []string `json:"tags" example:"Inspiratif,Programming"`
}

type BookNoteResponse struct {
	ID            string    `json:"id"`
	UserBookID    string    `json:"user_book_id"`
	Quote         string    `json:"quote"`
	PageReference int       `json:"page_reference"`
	Tags          []string  `json:"tags"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
