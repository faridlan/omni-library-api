package middleware

import (
	"time"

	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// GlobalLimiter membatasi request secara umum untuk melindungi server dari spam biasa.
// Aturan: 100 request per 1 menit untuk setiap IP.
func GlobalLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,             // Maksimal 100 request
		Expiration: 1 * time.Minute, // Dalam rentang waktu 1 menit
		LimitReached: func(c *fiber.Ctx) error {
			// Menggunakan fungsi SendError milikmu
			return utils.SendError(c, fiber.StatusTooManyRequests, "Terlalu banyak request. Silakan coba beberapa saat lagi.")
		},
	})
}

// StrictLimiter membatasi request dengan sangat ketat untuk endpoint yang rawan.
// Aturan: 5 request per 1 menit untuk setiap IP.
func StrictLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,               // Maksimal hanya 5 request
		Expiration: 1 * time.Minute, // Dalam rentang waktu 1 menit
		LimitReached: func(c *fiber.Ctx) error {
			// Menggunakan fungsi SendError milikmu
			return utils.SendError(c, fiber.StatusTooManyRequests, "Batas akses untuk aksi ini tercapai. Silakan tunggu 1 menit.")
		},
	})
}
