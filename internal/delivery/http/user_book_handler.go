package http

import (
	"github.com/faridlan/omni-library-api/internal/delivery/http/dto"
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UserBookHandler struct {
	usecase domain.UserBookUsecase
}

func NewUserBookHandler(u domain.UserBookUsecase) *UserBookHandler {
	return &UserBookHandler{usecase: u}
}

// ==========================================
// HELPER: MAPPING ENTITY KE RESPONSE DTO
// ==========================================
func toUserBookResponse(ub *domain.UserBook) dto.UserBookResponse {
	if ub == nil {
		return dto.UserBookResponse{}
	}
	return dto.UserBookResponse{
		ID:          ub.ID,
		UserID:      ub.UserID,
		BookID:      ub.BookID,
		Status:      ub.Status,
		CurrentPage: ub.CurrentPage,
		Rating:      ub.Rating,
		CreatedAt:   ub.CreatedAt,
		UpdatedAt:   ub.UpdatedAt,
	}
}

func toUserBookWithMetadataResponse(ub *domain.UserBookWithMetadata) dto.UserBookWithMetaDataResponse {
	if ub == nil {
		return dto.UserBookWithMetaDataResponse{}
	}

	return dto.UserBookWithMetaDataResponse{
		UserBookResponse: toUserBookResponse(&ub.UserBook),
		Book: dto.BookResponse{
			ID:            ub.Book.ID,
			ISBN:          ub.Book.ISBN,
			Title:         ub.Book.Title,
			Authors:       ub.Book.Authors,
			PublishedDate: ub.Book.PublishedDate,
			Description:   ub.Book.Description,
			PageCount:     ub.Book.PageCount,
			CoverURL:      ub.Book.CoverURL,
			CreatedAt:     ub.Book.CreatedAt,
			UpdatedAt:     ub.Book.UpdatedAt,
		},
	}
}

// AddBook godoc
// @Summary Tambah Buku ke Rak
// @Description Memasukkan buku dari database master ke dalam rak bacaan personal user (Default: TO_READ)
// @Tags Library
// @Accept json
// @Produce json
// @Param request body dto.AddBookRequest true "Payload berisi ID Buku"
// @Success 201 {object} utils.SuccessResponse[dto.UserBookResponse] "Buku berhasil ditambahkan ke rak"
// @Failure 400 {object} utils.ErrorResponse "Format JSON salah"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 409 {object} utils.ErrorResponse "Buku sudah ada di rak (Conflict)"
// @Failure 500 {object} utils.ErrorResponse "Gagal menyimpan buku ke rak"
// @Security BearerAuth
// @Router /api/library [post]
func (h *UserBookHandler) AddBook(c *fiber.Ctx) error {
	var req dto.AddBookRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	userID := c.Locals("user_id").(string)

	result, err := h.usecase.TrackNewBook(c.Context(), userID, req.BookID)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	// Gunakan Mapper!
	res := toUserBookResponse(result)

	return utils.SendSuccess(c, fiber.StatusCreated, "Buku berhasil ditambahkan ke rak", res)
}

// UpdateProgress godoc
// @Summary Update Progres Bacaan
// @Description Mengubah status, halaman saat ini, dan memberikan rating pada buku yang sedang dibaca
// @Tags Library
// @Accept json
// @Produce json
// @Param book_id path string true "ID Buku di database master"
// @Param request body dto.UpdateProgressRequest true "Payload update progres"
// @Success 200 {object} utils.SuccessResponse[dto.UserBookResponse] "Progres bacaan berhasil diperbarui"
// @Failure 400 {object} utils.ErrorResponse "Format JSON salah"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan di rak"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/library/{book_id} [put]
func (h *UserBookHandler) UpdateProgress(c *fiber.Ctx) error {
	bookID := c.Params("book_id")

	if err := utils.ValidateUUID(bookID, "book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	var req dto.UpdateProgressRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}
	userID := c.Locals("user_id").(string)

	// Note: Di kodingan sebelumnya kamu mengisi input.ID dengan userID.
	// Seharusnya input.ID dibiarkan kosong karena kita mencari berdasarkan userID dan bookID
	input := domain.UpdateUserBookInput{
		UserID: userID,
		BookID: bookID,
		Status: req.Status,
		Page:   req.CurrentPage,
		Rating: req.Rating,
	}

	result, err := h.usecase.UpdateReadingStatus(c.Context(), input)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	// Gunakan Mapper!
	res := toUserBookResponse(result)

	return utils.SendSuccess(c, fiber.StatusOK, "Progres bacaan berhasil diperbarui", res)
}

// GetMyLibrary godoc
// @Summary Lihat Isi Rak Buku
// @Description Menampilkan seluruh buku yang ada di rak personal user, lengkap dengan metadata bukunya. Bisa difilter berdasarkan status.
// @Tags Library
// @Produce json
// @Param page query int false "Nomor Halaman (Default: 1)"
// @Param limit query int false "Jumlah Data per Halaman (Default: 10)"
// @Param status query string false "Filter status: TO_READ, READING, FINISHED"
// @Success 200 {object} utils.PaginatedResponse[dto.UserBookWithMetaDataResponse] "Berhasil mengambil buku dari rak"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 404 {object} utils.ErrorResponse "Rak buku tidak ditemukan"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/library [get]
// @Security BearerAuth
func (h *UserBookHandler) GetMyLibrary(c *fiber.Ctx) error {
	statusFilter := c.Query("status")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	params := domain.PaginationQuery{
		Page:  page,
		Limit: limit,
	}

	userID := c.Locals("user_id").(string)

	books, meta, err := h.usecase.GetUserLibrary(c.Context(), userID, statusFilter, params)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	// MAPPING ARRAY
	var res []dto.UserBookWithMetaDataResponse
	for _, b := range books {
		res = append(res, toUserBookWithMetadataResponse(b))
	}

	if res == nil {
		res = make([]dto.UserBookWithMetaDataResponse, 0)
	}

	return utils.SendSuccessPaginated(c, "Berhasil mengambil buku dari rak", res, meta)
}

// GetUserBookDetail godoc
// @Summary Detail Buku di Rak
// @Description Menampilkan detail spesifik dari satu buku yang ada di rak pengguna.
// @Tags Library
// @Produce json
// @Param book_id path string true "ID Buku"
// @Success 200 {object} utils.SuccessResponse[dto.UserBookWithMetaDataResponse] "Berhasil mengambil detail buku"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan di rak"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/library/{book_id} [get]
// @Security BearerAuth
func (h *UserBookHandler) GetUserBookDetail(c *fiber.Ctx) error {
	bookID := c.Params("book_id")
	if err := utils.ValidateUUID(bookID, "book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}
	userID := c.Locals("user_id").(string)

	result, err := h.usecase.GetUserBookDetail(c.Context(), userID, bookID)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}
	if result == nil {
		return utils.SendError(c, fiber.StatusNotFound, "Buku tidak ditemukan di rak")
	}

	// Gunakan Mapper!
	res := toUserBookWithMetadataResponse(result)

	return utils.SendSuccess(c, fiber.StatusOK, "Berhasil mengambil detail buku dari rak", res)
}

// DeleteBookFromShelf godoc
// @Summary Hapus Buku dari Rak
// @Description Menghapus buku dari rak personal user
// @Tags Library
// @Produce json
// @Param book_id path string true "ID Buku"
// @Success 200 {object} utils.SuccessResponse[string] "Buku berhasil dihapus dari rak"
// @Failure 400 {object} utils.ErrorResponse "Format UUID salah"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan di rak"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/library/{book_id} [delete]
// @Security BearerAuth
func (h *UserBookHandler) DeleteBookFromShelf(c *fiber.Ctx) error {
	bookID := c.Params("book_id")
	if err := utils.ValidateUUID(bookID, "book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}
	userID := c.Locals("user_id").(string)

	err := h.usecase.DeleteBookFromShelf(c.Context(), userID, bookID)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Buku berhasil dihapus dari rak", nil)
}
