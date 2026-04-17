package http

import (
	"github.com/faridlan/omni-library-api/internal/delivery/http/middleware" // <-- Tambahkan import ini
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authUC domain.AuthUsecase, bookUC domain.BookUsecase, userBookUC domain.UserBookUsecase, noteUC domain.BookNoteUsecase) {
	api := app.Group("/api")

	// RUTE PUBLIK (Tanpa Satpam)
	NewAuthHandler(api, authUC)
	NewBookHandler(api, bookUC) // Asumsi: Lihat katalog buku bebas tanpa login

	// ==========================================
	// AREA VIP (Dilindungi Satpam JWT)
	// ==========================================
	// Kita buat grup baru khusus untuk rute yang butuh login
	protectedGroup := api.Group("/", middleware.Protected())

	// Daftarkan Handler yang butuh login ke grup VIP ini
	NewUserBookHandler(protectedGroup, userBookUC)
	NewBookNoteHandler(protectedGroup, noteUC)
}
