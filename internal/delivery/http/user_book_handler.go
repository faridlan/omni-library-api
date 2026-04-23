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

func NewUserBookHandler(router fiber.Router, u domain.UserBookUsecase) *UserBookHandler {
	return &UserBookHandler{usecase: u}

}

// AddBook godoc
// @Summary Tambah Buku ke Rak
// @Description Memasukkan buku dari database master ke dalam rak bacaan personal user (Default: TO_READ)
// @Tags Library
// @Accept json
// @Produce json
// @Param request body dto.AddBookRequest true "Payload berisi ID Buku"
// @Success 201 {object} domain.UserBook "Buku berhasil ditambahkan"
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

	return utils.SendSuccess(c, fiber.StatusCreated, "Buku berhasil ditambahkan ke rak", result)
}

// UpdateProgress godoc
// @Summary Update Progres Bacaan
// @Description Mengubah status, halaman saat ini, dan memberikan rating pada buku yang sedang dibaca
// @Tags Library
// @Accept json
// @Produce json
// @Param book_id path string true "ID Buku di database master"
// @Param request body dto.UpdateProgressRequest true "Payload update progres"
// @Success 200 {object} domain.UserBook "Berhasil update progres"
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

	result, err := h.usecase.UpdateReadingStatus(c.Context(), userID, bookID, req.Status, req.CurrentPage, req.Rating)
	if err != nil {

		return utils.HandleDomainError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Progres bacaan berhasil diperbarui", result)
}

// GetMyLibrary godoc
// @Summary Lihat Isi Rak Buku
// @Description Menampilkan seluruh buku yang ada di rak personal user, lengkap dengan metadata bukunya. Bisa difilter berdasarkan status.
// @Tags Library
// @Produce json
// @Param page query int false "Nomor Halaman (Default: 1)"
// @Param limit query int false "Jumlah Data per Halaman (Default: 10)"
// @Param status query string false "Filter status: TO_READ, READING, FINISHED"
// @Success 200 {array} utils.PaginatedResponse "Daftar buku di rak"
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

	if books == nil {
		books = make([]*domain.UserBookWithMetadata, 0)
	}

	return utils.SendSuccessPaginated(c, "Berhasil mengambil buku dari rak", books, meta)
}
