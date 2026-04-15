package http

// ==========================================
// DTO UNTUK REQUEST
// ==========================================

// FetchBookRequest adalah bentuk JSON yang kita harapkan dari Postman/Swagger
type FetchBookRequest struct {
	ISBN string `json:"isbn" example:"9786020633176" validate:"required"`
}

// ==========================================
// DTO UNTUK RESPONSE (Bisa kita siapkan juga untuk error nanti)
// ==========================================

// ErrorResponse adalah bentuk standar kalau API kita error
type ErrorResponse struct {
	Error  string `json:"error" example:"buku tidak ditemukan"`
	Detail string `json:"detail,omitempty"`
}
