package http

import (
	"github.com/faridlan/omni-library-api/internal/delivery/http/middleware"
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes adalah Peta Induk (Centralized Router) untuk seluruh aplikasi
func SetupRoutes(app *fiber.App, authUC domain.AuthUsecase, bookUC domain.BookUsecase, userBookUC domain.UserBookUsecase, noteUC domain.BookNoteUsecase) {
	// Grup Utama
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
	// (Tambahkan rute note lain di sini jika ada)

	// ==========================================
	// 🔴 KAWASAN ADMIN (Wajib Login + Role Admin)
	// ==========================================
	admin := protected.Group("/books", middleware.AdminOnly()) // Melanjutkan dari grup protected

	admin.Post("/manual", bookHandler.CreateManual)
	admin.Put("/:id", bookHandler.UpdateBook)
	admin.Delete("/:id", bookHandler.DeleteBook)
}
