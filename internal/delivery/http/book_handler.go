package http

import (
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type BookHandler struct {
	bookUsecase domain.BookUsecase
}

func NewBookHandler(router fiber.Router, bu domain.BookUsecase) *BookHandler {
	return &BookHandler{
		bookUsecase: bu,
	}

	// // Buat grup dasar untuk buku (akan menjadi /api/books)
	// bookGroup := router.Group("/books")

	// // 🟢 RUTE PUBLIK (Bebas tanpa login)
	// bookGroup.Get("/", handler.GetAll)

	// // 🟡 RUTE USER BIASA (Wajib login, tapi tidak harus admin)
	// // Satpam Protected() dipasang langsung spesifik di endpoint ini
	// bookGroup.Post("/fetch", middleware.Protected(), handler.FetchAndSave)

	// // 🔴 RUTE ADMIN (Wajib login + Wajib Admin)
	// // Kita buat sub-grup yang dijaga ketat oleh dua lapis Satpam
	// adminGroup := bookGroup.Group("/", middleware.Protected(), middleware.AdminOnly())

	// adminGroup.Post("/manual", handler.CreateManual)
	// adminGroup.Put("/:id", handler.UpdateBook)
	// adminGroup.Delete("/:id", handler.DeleteBook)
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
// @Security BearerAuth
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

// CreateManual godoc
// @Summary Tambah Buku Manual (Admin Only)
// @Description Menambahkan buku lokal tanpa ISBN ke database master
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} domain.Book
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Tidak bawa Token)"
// @Failure 403 {object} utils.ErrorResponse "Forbidden (Bukan Admin)"
// @Param request body BookRequest true "Payload data buku"
// @Router /api/books/manual [post]
func (h *BookHandler) CreateManual(c *fiber.Ctx) error {
	// Nanti kita panggil h.bookUsecase.CreateManual(...) di sini
	var req BookRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	var pubDate time.Time
	if req.PublishedDate != "" {
		parsed, err := time.Parse("2006-01-02", req.PublishedDate)
		if err != nil {
			return utils.SendError(c, fiber.StatusBadRequest, "Format published_date harus YYYY-MM-DD")
		}
		pubDate = parsed
	}

	// Ubah DTO menjadi Domain
	newBook := &domain.Book{
		ISBN:          req.ISBN,
		Title:         req.Title,
		Authors:       req.Authors,
		PublishedDate: pubDate,
		Description:   req.Description,
		PageCount:     req.PageCount,
		CoverURL:      req.CoverURL,
	}

	result, err := h.bookUsecase.CreateManual(c.Context(), newBook)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateBook godoc
// @Summary Update Data Buku (Admin Only)
// @Description Mengedit metadata buku master (judul, cover, dll)
// @Tags Books
// @Accept json
// @Produce json
// @Param id path string true "ID Buku"
// @Security BearerAuth
// @Success 200 {object} domain.Book
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Param request body BookRequest true "Payload data update"
// @Router /api/books/{id} [put]
func (h *BookHandler) UpdateBook(c *fiber.Ctx) error {
	bookID := c.Params("id")
	if err := utils.ValidateUUID(bookID, "book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	var req BookRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	var pubDate time.Time
	if req.PublishedDate != "" {
		parsed, err := time.Parse("2006-01-02", req.PublishedDate)
		if err != nil {
			return utils.SendError(c, fiber.StatusBadRequest, "Format published_date harus YYYY-MM-DD")
		}
		pubDate = parsed
	}

	updateData := &domain.Book{
		ISBN:          req.ISBN,
		Title:         req.Title,
		Authors:       req.Authors,
		PublishedDate: pubDate,
		Description:   req.Description,
		PageCount:     req.PageCount,
		CoverURL:      req.CoverURL,
	}

	result, err := h.bookUsecase.UpdateBook(c.Context(), bookID, updateData)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// DeleteBook godoc
// @Summary Hapus Buku (Admin Only)
// @Description Menghapus buku dari database master secara permanen
// @Tags Books
// @Produce json
// @Param id path string true "ID Buku"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Router /api/books/{id} [delete]
func (h *BookHandler) DeleteBook(c *fiber.Ctx) error {
	bookID := c.Params("id")
	if err := utils.ValidateUUID(bookID, "book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	err := h.bookUsecase.DeleteBook(c.Context(), bookID)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Buku berhasil dihapus dari sistem",
	})
}
