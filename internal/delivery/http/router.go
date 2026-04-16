package http

import (
	"github.com/faridlan/omni-library-api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes adalah resepsionis utama yang menghubungkan URL ke Handler yang tepat
func SetupRoutes(app *fiber.App, bookUC domain.BookUsecase, userBookUC domain.UserBookUsecase, noteUC domain.BookNoteUsecase) {
	// Grup utama untuk semua API
	api := app.Group("/api")

	// Panggil masing-masing constructor handler
	// Di sinilah fungsi NewBookHandler dkk mendaftarkan dirinya ke grup "/api"
	NewBookHandler(api, bookUC)
	NewUserBookHandler(api, userBookUC)
	NewBookNoteHandler(api, noteUC)
}
