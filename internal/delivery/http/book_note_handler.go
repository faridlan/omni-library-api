package http

import (
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type BookNoteHandler struct {
	usecase domain.BookNoteUsecase
}

func NewBookNoteHandler(router fiber.Router, u domain.BookNoteUsecase) {
	handler := &BookNoteHandler{usecase: u}

	// Kita buat sub-group di bawah URL yang butuh user_book_id
	noteGroup := router.Group("/library/:user_book_id/notes")

	noteGroup.Post("/", handler.AddNote)
	noteGroup.Get("/", handler.GetNotes)
}

func (h *BookNoteHandler) AddNote(c *fiber.Ctx) error {
	// Tangkap user_book_id dari URL
	userBookID := c.Params("user_book_id")

	type request struct {
		Quote         string   `json:"quote"`
		PageReference int      `json:"page_reference"`
		Tags          []string `json:"tags"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format JSON salah"})
	}

	// Susun data menjadi struct Domain
	note := &domain.BookNote{
		UserBookID:    userBookID,
		Quote:         req.Quote,
		PageReference: req.PageReference,
		Tags:          req.Tags, // Array string dari JSON langsung diteruskan
	}

	// Eksekusi lewat Usecase
	err := h.usecase.AddNote(c.Context(), note)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(note)
}

func (h *BookNoteHandler) GetNotes(c *fiber.Ctx) error {
	// Tangkap user_book_id dari URL
	userBookID := c.Params("user_book_id")

	notes, err := h.usecase.GetNotesForBook(c.Context(), userBookID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(notes)
}
