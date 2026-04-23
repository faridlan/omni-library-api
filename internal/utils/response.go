package utils

import (
	"errors"
	"log/slog"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Error  string `json:"error" example:"pesan error"`
	Detail string `json:"detail,omitempty"`
}

type PaginatedResponse struct {
	Message string                `json:"message"`
	Data    any                   `json:"data"`
	Meta    domain.PaginationMeta `json:"meta"`
}

type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // omitempty: jika nil, key "data" tidak ditampilkan (opsional)
}

// SendError adalah helper agar Handler kita makin tipis
func SendError(c *fiber.Ctx, statusCode int, message string, detail ...string) error {
	resp := ErrorResponse{
		Error: message,
	}

	// ====================================================================
	// FILTER KEAMANAN & STRUCTURED LOGGING
	// ====================================================================
	if statusCode >= fiber.StatusInternalServerError {

		// 1. Siapkan detail error asli (jika ada)
		var errDetail string
		if len(detail) > 0 && detail[0] != "" {
			errDetail = detail[0] // Isi variabelnya di sini
		}

		// 2. LOGGING: Panggil slog.Error SATU KALI saja
		slog.Error("CRITICAL SERVER ERROR",
			slog.Int("status_code", statusCode),
			slog.String("path", c.Path()),     // Catat URL mana yang error
			slog.String("method", c.Method()), // Catat methodnya (POST/GET)
			slog.String("error_message", message),
			slog.String("sql_detail", errDetail), // Otomatis terisi atau kosong
		)

		// 3. MASKING: Pastikan detail dikosongkan agar tidak masuk ke JSON Response
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

func SendSuccessPaginated(c *fiber.Ctx, message string, data any, meta domain.PaginationMeta) error {
	return c.Status(fiber.StatusOK).JSON(PaginatedResponse{
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func SendSuccess(c *fiber.Ctx, statusCode int, message string, data any) error {
	return c.Status(statusCode).JSON(SuccessResponse{
		Message: message,
		Data:    data,
	})
}
