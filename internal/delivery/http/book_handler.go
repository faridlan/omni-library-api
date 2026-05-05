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

func NewBookHandler(bu domain.BookUsecase) *BookHandler {
	return &BookHandler{
		bookUsecase: bu,
	}
}

// ==========================================
// HELPER: MAPPING ENTITY KE RESPONSE DTO
// ==========================================
func toBookResponse(b *domain.Book) dto.BookResponse {
	if b == nil {
		return dto.BookResponse{}
	}
	return dto.BookResponse{
		ID:            b.ID,
		ISBN:          b.ISBN,
		Title:         b.Title,
		Authors:       b.Authors,
		PublishedDate: b.PublishedDate,
		Description:   b.Description,
		PageCount:     b.PageCount,
		CoverURL:      b.CoverURL,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

// FetchAndSave godoc
// @Summary      Ambil & Simpan Metadata Buku
// @Description  Mencari buku di Google Books via ISBN dan menyimpannya ke database lokal
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        request body dto.FetchBookRequest true "Payload berisi ISBN"
// @Success      200 {object} utils.SuccessResponse[dto.BookResponse] "Metadata buku berhasil diambil"
// @Failure      400 {object} utils.ErrorResponse "Format JSON salah / Validasi gagal"
// @Failure      404 {object} utils.ErrorResponse "Buku tidak ditemukan di Google Books"
// @Failure      500 {object} utils.ErrorResponse "Internal Server Error"
// @Router       /api/books/fetch [post]
// @Security     BearerAuth
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

	// MAPPING KE RESPONSE DTO
	res := toBookResponse(book)

	return utils.SendSuccess(c, fiber.StatusOK, "Metadata buku berhasil diambil", res)
}

// GetAll godoc
// @Summary      Ambil Katalog Buku
// @Description  Mengambil daftar buku dari database secara terpaginasi (pagination)
// @Tags         Books
// @Produce      json
// @Param        page query int false "Nomor Halaman (Default: 1)"
// @Param        limit query int false "Jumlah Data per Halaman (Default: 10)"
// @Success      200 {object} utils.PaginatedResponse[dto.BookResponse] "Berhasil mengambil katalog buku"
// @Failure      500 {object} utils.ErrorResponse "Internal Server Error"
// @Router       /api/books [get]
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

	// MAPPING ARRAY ENTITY KE ARRAY RESPONSE DTO
	var res []dto.BookResponse
	for _, b := range books {
		res = append(res, toBookResponse(b))
	}
	// Pastikan array tidak null (menghindari kembalian "null" di JSON, diganti "[]")
	if res == nil {
		res = make([]dto.BookResponse, 0)
	}

	return utils.SendSuccessPaginated(c, "Berhasil mengambil katalog buku", res, meta)
}

// GetBookByID godoc
// @Summary      Ambil Buku Berdasarkan ID
// @Description  Mengambil detail buku berdasarkan ID yang diberikan
// @Tags         Books
// @Produce      json
// @Param        id path string true "ID Buku"
// @Success      200 {object} utils.SuccessResponse[dto.BookResponse] "Buku ditemukan"
// @Failure      404 {object} utils.ErrorResponse "Buku tidak ditemukan"
// @Router       /api/books/{id} [get]
func (h *BookHandler) GetBookByID(c *fiber.Ctx) error {
	bookID := c.Params("id")
	if err := utils.ValidateUUID(bookID, "book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	book, err := h.bookUsecase.GetBookByID(c.Context(), bookID)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	res := toBookResponse(book)

	return utils.SendSuccess(c, fiber.StatusOK, "Buku ditemukan", res)
}

// CreateManual godoc
// @Summary      Tambah Buku Manual (Admin Only)
// @Description  Menambahkan buku lokal tanpa ISBN ke database master
// @Tags         Books
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.BookRequest true "Payload data buku"
// @Success      201 {object} utils.SuccessResponse[dto.BookResponse] "Buku berhasil ditambahkan"
// @Failure      400 {object} utils.ErrorResponse "Format JSON salah / Validasi gagal"
// @Failure      401 {object} utils.ErrorResponse "Unauthorized"
// @Failure      403 {object} utils.ErrorResponse "Forbidden"
// @Failure      409 {object} utils.ErrorResponse "ISBN sudah ada"
// @Router       /api/books/manual [post]
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

	// MAPPING DTO -> DOMAIN INPUT (Sesuai kritik arsitektur sebelumnya)
	input := domain.CreateBookInput{
		ISBN:          req.ISBN,
		Title:         req.Title,
		Authors:       req.Authors,
		PublishedDate: pubDate,
		Description:   req.Description,
		PageCount:     req.PageCount,
		CoverURL:      req.CoverURL,
	}

	// Panggil Usecase (Pastikan Usecase-mu sudah menggunakan domain.CreateBookInput)
	result, err := h.bookUsecase.CreateManual(c.Context(), input)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	// MAPPING BALIK KE RESPONSE DTO
	res := toBookResponse(result)

	return utils.SendSuccess(c, fiber.StatusCreated, "Buku berhasil ditambahkan", res)
}

// UpdateBook godoc
// @Summary      Update Data Buku (Admin Only)
// @Description  Mengedit metadata buku master (judul, cover, dll)
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        id path string true "ID Buku"
// @Param        request body dto.BookRequest true "Payload data update"
// @Security     BearerAuth
// @Success      200 {object} utils.SuccessResponse[dto.BookResponse] "Metadata buku berhasil diperbarui"
// @Failure      400 {object} utils.ErrorResponse "Bad Request"
// @Failure      401 {object} utils.ErrorResponse "Unauthorized"
// @Failure      403 {object} utils.ErrorResponse "Forbidden"
// @Router       /api/books/{id} [put]
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

	// MAPPING DTO -> DOMAIN INPUT
	input := domain.UpdateBookInput{
		ID:            bookID,
		ISBN:          req.ISBN,
		Title:         req.Title,
		Authors:       req.Authors,
		PublishedDate: pubDate,
		Description:   req.Description,
		PageCount:     req.PageCount,
		CoverURL:      req.CoverURL,
	}

	// Panggil Usecase (Pastikan Usecase-mu sudah menggunakan domain.UpdateBookInput)
	result, err := h.bookUsecase.UpdateBook(c.Context(), input)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	res := toBookResponse(result)

	return utils.SendSuccess(c, fiber.StatusOK, "Metadata buku berhasil diperbarui", res)
}

// DeleteBook godoc
// @Summary      Hapus Buku (Admin Only)
// @Description  Menghapus buku dari database master secara permanen
// @Tags         Books
// @Produce      json
// @Param        id path string true "ID Buku"
// @Security     BearerAuth
// @Success      200 {object} utils.SuccessResponse[interface{}] "Buku berhasil dihapus dari sistem"
// @Failure      400 {object} utils.ErrorResponse "Format ID salah"
// @Failure      401 {object} utils.ErrorResponse "Unauthorized"
// @Failure      403 {object} utils.ErrorResponse "Forbidden"
// @Failure      404 {object} utils.ErrorResponse "Buku tidak ditemukan"
// @Router       /api/books/{id} [delete]
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
