package http

import (
	"github.com/faridlan/omni-library-api/internal/delivery/http/middleware"
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"

	// 1. Import library Prometheus untuk Fiber
	"github.com/ansrivas/fiberprometheus/v2"
)

// SetupRoutes adalah Peta Induk (Centralized Router) untuk seluruh aplikasi
func SetupRoutes(app *fiber.App, authUC domain.AuthUsecase, bookUC domain.BookUsecase, userBookUC domain.UserBookUsecase, noteUC domain.BookNoteUsecase) {

	// ==========================================
	// 📊 PROMETHEUS METRICS (Observability)
	// ==========================================
	// Inisialisasi Prometheus dengan nama aplikasi kita
	prometheus := fiberprometheus.New("omni_api")

	// Daftarkan endpoint rahasia di /metrics (Bisa diakses tanpa Token JWT)
	prometheus.RegisterAt(app, "/metrics")

	// Pasang middleware di root app agar mencatat SEMUA traffic yang lewat
	app.Use(prometheus.Middleware)

	// ==========================================
	// Grup Utama
	// ==========================================
	api := app.Group("/api")

	// 1. Inisialisasi Semua Handler
	authHandler := NewAuthHandler(api, authUC)
	bookHandler := NewBookHandler(api, bookUC)
	userBookHandler := NewUserBookHandler(api, userBookUC)
	bookNoteHandler := NewBookNoteHandler(api, noteUC)

	// ==========================================
	// 🟢 KAWASAN PUBLIK (Tanpa Satpam)
	// ==========================================

	// Auth
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	// Katalog Buku
	api.Get("/books", bookHandler.GetAll)
	api.Get("/books/:id", bookHandler.GetBookByID)

	// Endpoint Tracer Bullet untuk mengetes CI/CD
	api.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Hello dari Staging! CI/CD Otomatis berhasil mendarat dengan mulus. 🚀",
			"version": "1.0.1-beta",
		})
	})

	// ==========================================
	// 🟡 KAWASAN VIP (Wajib Login / Token JWT)
	// ==========================================
	protected := api.Group("/", middleware.Protected())

	// Aksi Buku (User)
	protected.Post("/books/fetch", bookHandler.FetchAndSave)

	// Rak Buku Personal (Library)
	lib := protected.Group("/library")
	lib.Post("/", userBookHandler.AddBook)
	lib.Get("/", userBookHandler.GetMyLibrary)
	lib.Put("/:book_id", userBookHandler.UpdateProgress)

	// Catatan Buku (Notes)
	notes := protected.Group("/library/:user_book_id/notes")
	notes.Post("/", bookNoteHandler.AddNote)
	notes.Get("/", bookNoteHandler.GetNotes)

	// ==========================================
	// 🔴 KAWASAN ADMIN (Wajib Login + Role Admin)
	// ==========================================
	admin := protected.Group("/books", middleware.AdminOnly()) // Melanjutkan dari grup protected

	admin.Post("/manual", bookHandler.CreateManual)
	admin.Put("/:id", bookHandler.UpdateBook)
	admin.Delete("/:id", bookHandler.DeleteBook)
}
