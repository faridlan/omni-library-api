package http

import (
	"github.com/faridlan/omni-library-api/internal/delivery/http/dto"
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type BookNoteHandler struct {
	usecase domain.BookNoteUsecase
}

func NewBookNoteHandler(router fiber.Router, u domain.BookNoteUsecase) *BookNoteHandler {
	return &BookNoteHandler{usecase: u}

}

// AddNote godoc
// @Summary Tambah Kutipan/Catatan Buku
// @Description Menyimpan kutipan favorit, referensi halaman, dan tag untuk buku tertentu yang ada di rak
// @Tags Book Notes
// @Accept json
// @Produce json
// @Param user_book_id path string true "ID progres buku di rak (Bukan master Book ID)"
// @Param request body dto.AddNoteRequest true "Payload isi kutipan dan tag"
// @Success 200 {object} utils.SuccessResponse[domain.BookNote] "Note buku berhasil ditambahkan"
// @Failure 400 {object} utils.ErrorResponse "Format JSON salah atau Quote kosong"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan di rak"
// @Failure 500 {object} utils.ErrorResponse "Gagal menyimpan catatan"
// @Security BearerAuth
// @Router /api/library/{user_book_id}/notes [post]
func (h *BookNoteHandler) AddNote(c *fiber.Ctx) error {
	userBookID := c.Params("user_book_id")

	if err := utils.ValidateUUID(userBookID, "user_book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	var req dto.AddNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	note := &domain.BookNote{
		UserBookID:    userBookID,
		Quote:         req.Quote,
		PageReference: req.PageReference,
		Tags:          req.Tags,
	}

	err := h.usecase.AddNote(c.Context(), note)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Note buku berhasil ditambahkan", note)
}

// GetNotes godoc
// @Summary Ambil Daftar Catatan Buku
// @Description Melihat seluruh kutipan dan catatan yang pernah ditulis untuk satu buku spesifik di rak
// @Tags Book Notes
// @Produce json
// @Param page query int false "Nomor Halaman (Default: 1)"
// @Param limit query int false "Jumlah Data per Halaman (Default: 10)"
// @Param user_book_id path string true "ID progres buku di rak"
// @Success 200 {object} utils.PaginatedResponse[domain.BookNote] "Berhasil mengambil note buku"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan di rak"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/library/{user_book_id}/notes [get]
func (h *BookNoteHandler) GetNotes(c *fiber.Ctx) error {
	userBookID := c.Params("user_book_id")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	params := domain.PaginationQuery{
		Page:  page,
		Limit: limit,
	}

	if err := utils.ValidateUUID(userBookID, "user_book_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	notes, meta, err := h.usecase.GetNotesForBook(c.Context(), userBookID, params)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	if notes == nil {
		notes = make([]*domain.BookNote, 0)
	}

	return utils.SendSuccessPaginated(c, "Berhasil mengambil note buku", notes, meta)
}

// DeleteNote godoc
// @Summary Hapus Catatan Buku
// @Description Menghapus satu catatan buku spesifik
// @Tags Book Notes
// @Produce json
// @Param user_book_id path string true "ID progres buku di rak"
// @Param note_id path string true "ID catatan buku"
// @Success 200 {object} utils.SuccessResponse[utils.EmptyObj] "Note buku berhasil dihapus"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 404 {object} utils.ErrorResponse "Note buku tidak ditemukan"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/library/{user_book_id}/notes/{note_id} [delete]
func (h *BookNoteHandler) DeleteNote(c *fiber.Ctx) error {
	noteID := c.Params("note_id")

	if err := utils.ValidateUUID(noteID, "note_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	err := h.usecase.DeleteNote(c.Context(), noteID)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Note buku berhasil dihapus", nil)
}

// UpdateNote godoc
// @Summary Perbarui Catatan Buku
// @Description Memperbarui informasi satu catatan buku spesifik
// @Tags Book Notes
// @Produce json
// @Param user_book_id path string true "ID progres buku di rak"
// @Param note_id path string true "ID catatan buku"
// @Success 200 {object} utils.SuccessResponse[domain.BookNote] "Note buku berhasil diperbarui"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 404 {object} utils.ErrorResponse "Note buku tidak ditemukan"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Param request body dto.UpdateNoteRequest true "Payload data update"
// @Router /api/library/{user_book_id}/notes/{note_id} [put]
func (h *BookNoteHandler) UpdateNote(c *fiber.Ctx) error {
	noteID := c.Params("note_id")

	if err := utils.ValidateUUID(noteID, "note_id"); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	var req dto.UpdateNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	req.ID = noteID

	note := &domain.BookNote{
		ID:            req.ID,
		Quote:         req.Quote,
		PageReference: req.PageReference,
		Tags:          req.Tags,
	}

	updatedNote, err := h.usecase.UpdateNote(c.Context(), note)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Note buku berhasil diperbarui", updatedNote)
}
