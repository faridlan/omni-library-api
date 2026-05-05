package http

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/faridlan/omni-library-api/internal/delivery/http/middleware"
	"github.com/gofiber/fiber/v2"
)

type AppHandlers struct {
	Auth     *AuthHandler
	Book     *BookHandler
	UserBook *UserBookHandler
	BookNote *BookNoteHandler
	User     *UserHandler
}

func SetupRoutes(app *fiber.App, h AppHandlers) {

	prometheus := fiberprometheus.New("omni_api")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	api := app.Group("/api")

	// ==========================================
	// PUBLIC ROUTES
	// ==========================================
	auth := api.Group("/auth")
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
	auth.Post("/refresh", h.Auth.Refresh)

	api.Get("/books", h.Book.GetAll)
	api.Get("/books/:id", h.Book.GetBookByID)

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Hello dari Staging! CI/CD Otomatis berhasil mendarat dengan mulus. 🚀",
			"version": "1.0.1-beta",
		})
	})

	// ==========================================
	// PROTECTED ROUTES (Butuh Login)
	// ==========================================
	protected := api.Group("/", middleware.Protected())

	protected.Post("/books/fetch", h.Book.FetchAndSave)

	lib := protected.Group("/library")
	lib.Post("/", h.UserBook.AddBook)
	lib.Get("/", h.UserBook.GetMyLibrary)
	lib.Put("/:book_id", h.UserBook.UpdateProgress)
	lib.Delete("/:book_id", h.UserBook.DeleteBookFromShelf)
	lib.Get("/:book_id", h.UserBook.GetUserBookDetail)

	notes := protected.Group("/library/:user_book_id/notes")
	notes.Post("/", h.BookNote.AddNote)
	notes.Get("/", h.BookNote.GetNotes)
	notes.Delete("/:note_id", h.BookNote.DeleteNote)
	notes.Put("/:note_id", h.BookNote.UpdateNote)

	userGroup := protected.Group("/users")
	userGroup.Get("/me", h.User.GetProfile)
	userGroup.Put("/me", h.User.UpdateProfile)
	userGroup.Put("/me/password", h.User.UpdatePassword)

	// ==========================================
	// ADMIN ROUTES
	// ==========================================
	admin := protected.Group("/books", middleware.AdminOnly())
	admin.Post("/manual", h.Book.CreateManual)
	admin.Put("/:id", h.Book.UpdateBook)
	admin.Delete("/:id", h.Book.DeleteBook)
}
