package http

import (
	"time"

	"github.com/faridlan/omni-library-api/internal/delivery/http/dto"
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
}

// FetchAndSave godoc
// @Summary Ambil & Simpan Metadata Buku
// @Description Mencari buku di Google Books via ISBN dan menyimpannya ke database lokal
// @Tags Books
// @Accept json
// @Produce json
// @Param request body dto.FetchBookRequest true "Payload berisi ISBN"
// @Success 200 {object} utils.SuccessResponse[domain.Book] "Metadata buku berhasil diambil"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan di Google Books"
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/books/fetch [post]
// @Security BearerAuth
func (h *BookHandler) FetchAndSave(c *fiber.Ctx) error {

	var req dto.FetchBookRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	book, err := h.bookUsecase.FetchAndSaveMetadata(c.Context(), req.ISBN)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Metadata buku berhasil diambil", book)
}

// GetAll godoc
// @Summary Ambil Katalog Buku
// @Description Mengambil daftar buku dari database secara terpaginasi (pagination)
// @Tags Books
// @Produce json
// @Param page query int false "Nomor Halaman (Default: 1)"
// @Param limit query int false "Jumlah Data per Halaman (Default: 10)"
// @Success 200 {object} utils.PaginatedResponse[domain.Book] "Berhasil mengambil katalog buku"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/books [get]
func (h *BookHandler) GetAll(c *fiber.Ctx) error {

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	params := domain.PaginationQuery{
		Page:  page,
		Limit: limit,
	}

	books, meta, err := h.bookUsecase.GetAllBooks(c.Context(), params)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data buku", err.Error())
	}

	if books == nil {
		books = make([]*domain.Book, 0)
	}

	return utils.SendSuccessPaginated(c, "Berhasil mengambil katalog buku", books, meta)
}

// GetBookByID godoc
// @Summary Ambil Buku Berdasarkan ID
// @Description Mengambil detail buku berdasarkan ID yang diberikan
// @Tags Books
// @Produce json
// @Param id path string true "ID Buku"
// @Success 200 {object} utils.SuccessResponse[domain.Book] "Buku ditemukan"
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan"
// @Router /api/books/{id} [get]
func (h *BookHandler) GetBookByID(c *fiber.Ctx) error {
	bookID := c.Params("id")
	if err := utils.ValidateUUID(bookID, "book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	book, err := h.bookUsecase.GetBookByID(c.Context(), bookID)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Buku ditemukan", book)
}

// CreateManual godoc
// @Summary Tambah Buku Manual (Admin Only)
// @Description Menambahkan buku lokal tanpa ISBN ke database master
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} utils.SuccessResponse[domain.Book] "Buku berhasil ditambahkan"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Tidak bawa Token)"
// @Failure 403 {object} utils.ErrorResponse "Forbidden (Bukan Admin)"
// @Param request body dto.BookRequest true "Payload data buku"
// @Router /api/books/manual [post]
func (h *BookHandler) CreateManual(c *fiber.Ctx) error {
	var req dto.BookRequest
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

	return utils.SendSuccess(c, fiber.StatusCreated, "Buku berhasil ditambahkan", result)
}

// UpdateBook godoc
// @Summary Update Data Buku (Admin Only)
// @Description Mengedit metadata buku master (judul, cover, dll)
// @Tags Books
// @Accept json
// @Produce json
// @Param id path string true "ID Buku"
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse[domain.Book] "Metadata buku berhasil diperbarui"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Param request body dto.BookRequest true "Payload data update"
// @Router /api/books/{id} [put]
func (h *BookHandler) UpdateBook(c *fiber.Ctx) error {
	bookID := c.Params("id")
	if err := utils.ValidateUUID(bookID, "book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	var req dto.BookRequest
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

	return utils.SendSuccess(c, fiber.StatusOK, "Metadata buku berhasil diperbarui", result)
}

// DeleteBook godoc
// @Summary Hapus Buku (Admin Only)
// @Description Menghapus buku dari database master secara permanen
// @Tags Books
// @Produce json
// @Param id path string true "ID Buku"
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse[utils.EmptyObj] "Buku berhasil dihapus dari sistem"
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

	return utils.SendSuccess(c, fiber.StatusOK, "Buku berhasil dihapus dari sistem", nil)
}
