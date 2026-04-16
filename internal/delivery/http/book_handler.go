package http

import (
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/utils"
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
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan di Google Books"
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/books/fetch [post]
func (h *BookHandler) FetchAndSave(c *fiber.Ctx) error {

	var req FetchBookRequest

	// 1. Tangkap JSON
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}

	// 2. VALIDASI OTOMATIS! (Membaca tag validate:"required" di DTO)
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	book, err := h.bookUsecase.FetchAndSaveMetadata(c.Context(), req.ISBN)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(book)
}

// GetAll godoc
// @Summary Ambil Katalog Buku
// @Description Mengambil daftar seluruh buku yang tersimpan di database lokal
// @Tags Books
// @Produce json
// @Success 200 {array} domain.Book "Berhasil mengambil daftar buku"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/books [get]
func (h *BookHandler) GetAll(c *fiber.Ctx) error {
	books, err := h.bookUsecase.GetAllBooks(c.Context())
	if err != nil {
		// Kita gunakan utils.SendError agar format JSON-nya konsisten 100%
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data buku", err.Error())
	}

	if books == nil {
		books = make([]*domain.Book, 0)
	}

	return c.Status(fiber.StatusOK).JSON(books)
}
