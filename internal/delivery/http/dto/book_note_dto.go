package dto

type AddNoteRequest struct {
	Quote         string   `json:"quote" example:"Bekerjalah seperti programmer pemalas..." validate:"required"`
	PageReference int      `json:"page_reference" example:"42"`
	Tags          []string `json:"tags" example:"Inspiratif,Programming"`
}

type UpdateNoteRequest struct {
	ID            string
	Quote         string   `json:"quote" example:"Bekerjalah seperti programmer pemalas..." validate:"required"`
	PageReference int      `json:"page_reference" example:"42"`
	Tags          []string `json:"tags" example:"Inspiratif,Programming"`
}
