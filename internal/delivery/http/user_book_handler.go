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
	libGroup.Get("/", handler.GetMyLibrary)
}

// ⚠️ HARDCODE SEMENTARA (Ganti dengan UUID dari database-mu)
const DummyUserID = "08a2fccf-46c8-473f-bb86-53a1c0b2b8a6"

// AddBook godoc
// @Summary Tambah Buku ke Rak
// @Description Memasukkan buku dari database master ke dalam rak bacaan personal user (Default: TO_READ)
// @Tags Library
// @Accept json
// @Produce json
// @Param request body AddBookRequest true "Payload berisi ID Buku"
// @Success 201 {object} domain.UserBook "Buku berhasil ditambahkan"
// @Failure 400 {object} ErrorResponse "Format JSON salah"
// @Failure 409 {object} ErrorResponse "Buku sudah ada di rak (Conflict)"
// @Router /api/library [post]
func (h *UserBookHandler) AddBook(c *fiber.Ctx) error {
	var req AddBookRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Format JSON salah"})
	}

	result, err := h.usecase.TrackNewBook(c.Context(), DummyUserID, req.BookID)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateProgress godoc
// @Summary Update Progres Bacaan
// @Description Mengubah status, halaman saat ini, dan memberikan rating pada buku yang sedang dibaca
// @Tags Library
// @Accept json
// @Produce json
// @Param book_id path string true "ID Buku di database master"
// @Param request body UpdateProgressRequest true "Payload update progres"
// @Success 200 {object} domain.UserBook "Berhasil update progres"
// @Failure 400 {object} ErrorResponse "Format JSON salah"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /api/library/{book_id} [put]
func (h *UserBookHandler) UpdateProgress(c *fiber.Ctx) error {
	bookID := c.Params("book_id")

	var req UpdateProgressRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Format JSON salah"})
	}

	result, err := h.usecase.UpdateReadingStatus(c.Context(), DummyUserID, bookID, req.Status, req.CurrentPage, req.Rating)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// GetMyLibrary godoc
// @Summary Lihat Isi Rak Buku
// @Description Menampilkan seluruh buku yang ada di rak personal user, lengkap dengan metadata bukunya. Bisa difilter berdasarkan status.
// @Tags Library
// @Produce json
// @Param status query string false "Filter status: TO_READ, READING, FINISHED"
// @Success 200 {array} domain.UserBookWithMetadata "Daftar buku di rak"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /api/library [get]
func (h *UserBookHandler) GetMyLibrary(c *fiber.Ctx) error {
	statusFilter := c.Query("status")

	books, err := h.usecase.GetUserLibrary(c.Context(), DummyUserID, statusFilter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	// Best practice: Cegah return JSON null jika rak masih kosong
	if books == nil {
		books = make([]*domain.UserBookWithMetadata, 0)
	}

	return c.Status(fiber.StatusOK).JSON(books)
}
