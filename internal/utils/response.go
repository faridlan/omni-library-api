package utils

import "github.com/gofiber/fiber/v2"

// ErrorResponse adalah standar DTO untuk semua error di aplikasi
type ErrorResponse struct {
	Error  string `json:"error" example:"pesan error"`
	Detail string `json:"detail,omitempty"`
}

// SendError adalah helper agar Handler kita makin tipis
func SendError(c *fiber.Ctx, statusCode int, message string, detail ...string) error {
	resp := ErrorResponse{
		Error: message,
	}

	// Jika ada detail error tambahan (misal dari GORM), kita masukkan
	if len(detail) > 0 && detail[0] != "" {
		resp.Detail = detail[0]
	}

	return c.Status(statusCode).JSON(resp)
}
