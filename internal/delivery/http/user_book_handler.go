package http

import (
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type UserBookHandler struct {
	usecase domain.UserBookUsecase
}

func NewUserBookHandler(router fiber.Router, u domain.UserBookUsecase) {
	handler := &UserBookHandler{usecase: u}

	// Grup endpoint khusus rak buku (library)
	libGroup := router.Group("/library")
	libGroup.Post("/", handler.AddBook)
	libGroup.Put("/:book_id", handler.UpdateProgress)
}

// ⚠️ HARDCODE SEMENTARA (Ganti dengan UUID dari database-mu)
const DummyUserID = "08a2fccf-46c8-473f-bb86-53a1c0b2b8a6"

func (h *UserBookHandler) AddBook(c *fiber.Ctx) error {
	type request struct {
		BookID string `json:"book_id"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format JSON salah"})
	}

	// Panggil Usecase menggunakan DummyUserID
	result, err := h.usecase.TrackNewBook(c.Context(), DummyUserID, req.BookID)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *UserBookHandler) UpdateProgress(c *fiber.Ctx) error {
	bookID := c.Params("book_id") // Ambil book_id dari URL (/:book_id)

	type request struct {
		Status      string `json:"status"`
		CurrentPage int    `json:"current_page"`
		Rating      int    `json:"rating"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format JSON salah"})
	}

	result, err := h.usecase.UpdateReadingStatus(c.Context(), DummyUserID, bookID, req.Status, req.CurrentPage, req.Rating)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
