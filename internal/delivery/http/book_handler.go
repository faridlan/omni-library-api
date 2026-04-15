package http

import (
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type BookHandler struct {
	bookUsecase domain.BookUsecase
}

func NewBookHandler(router fiber.Router, bu domain.BookUsecase) {
	handler := &BookHandler{
		bookUsecase: bu,
	}

	router.Post("/books/fetch", handler.FetchAndSave)
	router.Get("/books", handler.GetAll)
}

func (h *BookHandler) FetchAndSave(c *fiber.Ctx) error {
	// Request body sederhana
	type request struct {
		ISBN string `json:"isbn"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.ISBN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ISBN is required"})
	}

	book, err := h.bookUsecase.FetchAndSaveMetadata(c.Context(), req.ISBN)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(book)
}

func (h *BookHandler) GetAll(c *fiber.Ctx) error {
	// Memanggil Usecase
	books, err := h.bookUsecase.GetAllBooks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Gagal mengambil data buku",
			"detail": err.Error(),
		})
	}

	// Fiber secara otomatis akan mengonversi slice []*domain.Book menjadi JSON Array
	return c.Status(fiber.StatusOK).JSON(books)
}
