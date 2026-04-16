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

// FetchAndSave godoc
// @Summary Ambil & Simpan Metadata Buku
// @Description Mencari buku di Google Books via ISBN dan menyimpannya ke database lokal
// @Tags Books
// @Accept json
// @Produce json
// @Param request body FetchBookRequest true "Payload berisi ISBN"
// @Success 200 {object} domain.Book
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/books/fetch [post]
func (h *BookHandler) FetchAndSave(c *fiber.Ctx) error {

	var req FetchBookRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	if req.ISBN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "ISBN is required"})
	}

	book, err := h.bookUsecase.FetchAndSaveMetadata(c.Context(), req.ISBN)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(book)
}

// GetAll godoc
// @Summary Ambil Katalog Buku
// @Description Mengambil daftar seluruh buku yang tersimpan di database lokal
// @Tags Books
// @Produce json
// @Success 200 {array} domain.Book "Berhasil mengambil daftar buku"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /api/books [get]
func (h *BookHandler) GetAll(c *fiber.Ctx) error {

	books, err := h.bookUsecase.GetAllBooks(c.Context())
	if err != nil {
		// Menggunakan ErrorResponse DTO agar konsisten dengan Swagger
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:  "Gagal mengambil data buku",
			Detail: err.Error(),
		})
	}

	if books == nil {
		books = make([]*domain.Book, 0)
	}

	return c.Status(fiber.StatusOK).JSON(books)
}
