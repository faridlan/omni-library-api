package http

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/faridlan/omni-library-api/internal/delivery/http/middleware"
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authUC domain.AuthUsecase, bookUC domain.BookUsecase, userBookUC domain.UserBookUsecase, noteUC domain.BookNoteUsecase) {

	prometheus := fiberprometheus.New("omni_api")

	prometheus.RegisterAt(app, "/metrics")

	app.Use(prometheus.Middleware)

	api := app.Group("/api")

	authHandler := NewAuthHandler(api, authUC)
	bookHandler := NewBookHandler(api, bookUC)
	userBookHandler := NewUserBookHandler(api, userBookUC)
	bookNoteHandler := NewBookNoteHandler(api, noteUC)

	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	api.Get("/books", bookHandler.GetAll)
	api.Get("/books/:id", bookHandler.GetBookByID)

	api.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Hello dari Staging! CI/CD Otomatis berhasil mendarat dengan mulus. 🚀",
			"version": "1.0.1-beta",
		})
	})

	protected := api.Group("/", middleware.Protected())

	protected.Post("/books/fetch", bookHandler.FetchAndSave)

	lib := protected.Group("/library")
	lib.Post("/", userBookHandler.AddBook)
	lib.Get("/", userBookHandler.GetMyLibrary)
	lib.Put("/:book_id", userBookHandler.UpdateProgress)
	lib.Delete("/:book_id", userBookHandler.DeleteBookFromShelf)
	lib.Get("/:book_id", userBookHandler.GetUserBookDetail)

	notes := protected.Group("/library/:user_book_id/notes")
	notes.Post("/", bookNoteHandler.AddNote)
	notes.Get("/", bookNoteHandler.GetNotes)
	notes.Delete("/:note_id", bookNoteHandler.DeleteNote)
	notes.Put("/:note_id", bookNoteHandler.UpdateNote)

	admin := protected.Group("/books", middleware.AdminOnly())

	admin.Post("/manual", bookHandler.CreateManual)
	admin.Put("/:id", bookHandler.UpdateBook)
	admin.Delete("/:id", bookHandler.DeleteBook)
}
