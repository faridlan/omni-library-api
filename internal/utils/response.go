package utils

import (
	"errors"
	"log"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Error  string `json:"error" example:"pesan error"`
	Detail string `json:"detail,omitempty"`
}

// SendError adalah helper agar Handler kita makin tipis
func SendError(c *fiber.Ctx, statusCode int, message string, detail ...string) error {
	resp := ErrorResponse{
		Error: message,
	}

	// ====================================================================
	// FILTER KEAMANAN (Mencegah Information Disclosure)
	// ====================================================================
	if statusCode >= fiber.StatusInternalServerError {
		// 1. LOGGING: Cetak error asli ke terminal agar mudah di-debug
		if len(detail) > 0 && detail[0] != "" {
			log.Printf("[CRITICAL SERVER ERROR %d] %s: %s\n", statusCode, message, detail[0])
		} else {
			log.Printf("[CRITICAL SERVER ERROR %d] %s\n", statusCode, message)
		}

		// 2. MASKING: Pastikan detail dikosongkan agar tidak menjadi JSON
		resp.Detail = ""
	} else {
		// Untuk error 4xx (Client Error), aman untuk menampilkan detail ke user
		if len(detail) > 0 && detail[0] != "" {
			resp.Detail = detail[0]
		}
	}

	return c.Status(statusCode).JSON(resp)
}

// HandleDomainError menerjemahkan error dari Domain menjadi HTTP Status Code yang tepat
func HandleDomainError(c *fiber.Ctx, err error) error {
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
		// Kode aslimu TETAP DIPERTAHANKAN!
		// err.Error() tetap dikirimkan ke SendError.
		// Namun sekarang SendError akan mencetaknya ke terminal dan mencegahnya masuk ke JSON.
		return SendError(c, fiber.StatusInternalServerError, domain.ErrInternalServerError.Error(), err.Error())
	}
}
