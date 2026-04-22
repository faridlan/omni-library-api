package dto

// ==========================================
// DTO UNTUK REQUEST BOOK NOTES
// ==========================================

// AddNoteRequest adalah payload untuk menambahkan kutipan buku
type AddNoteRequest struct {
	Quote         string   `json:"quote" example:"Bekerjalah seperti programmer pemalas..." validate:"required"`
	PageReference int      `json:"page_reference" example:"42"`
	Tags          []string `json:"tags" example:"Inspiratif,Programming"`
}
