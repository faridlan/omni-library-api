package http

// ==========================================
// DTO UNTUK REQUEST
// ==========================================

// FetchBookRequest adalah bentuk JSON yang kita harapkan dari Postman/Swagger
type FetchBookRequest struct {
	ISBN string `json:"isbn" example:"9786020633176" validate:"required"`
}
