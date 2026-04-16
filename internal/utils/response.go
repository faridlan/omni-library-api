package utils

import (
	"errors"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

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

// HandleDomainError menerjemahkan error dari Domain menjadi HTTP Status Code yang tepat
func HandleDomainError(c *fiber.Ctx, err error) error {
	// Gunakan errors.Is() untuk mencocokkan tipe error
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return SendError(c, fiber.StatusNotFound, err.Error())

	case errors.Is(err, domain.ErrConflict):
		return SendError(c, fiber.StatusConflict, err.Error())

	case errors.Is(err, domain.ErrBadParamInput):
		return SendError(c, fiber.StatusBadRequest, err.Error())

	case errors.Is(err, domain.ErrLimitExceeded):
		return SendError(c, fiber.StatusTooManyRequests, err.Error())

	default:
		// Jika error tidak dikenali, berarti ada crash/bug di sistem (500)
		return SendError(c, fiber.StatusInternalServerError, domain.ErrInternalServerError.Error(), err.Error())
	}
}
