package http

import (
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type BookNoteHandler struct {
	usecase domain.BookNoteUsecase
}

func NewBookNoteHandler(router fiber.Router, u domain.BookNoteUsecase) {
	handler := &BookNoteHandler{usecase: u}

	// Kita buat sub-group di bawah URL yang butuh user_book_id
	noteGroup := router.Group("/library/:user_book_id/notes")

	noteGroup.Post("/", handler.AddNote)
	noteGroup.Get("/", handler.GetNotes)
}

// AddNote godoc
// @Summary Tambah Kutipan/Catatan Buku
// @Description Menyimpan kutipan favorit, referensi halaman, dan tag untuk buku tertentu yang ada di rak
// @Tags Book Notes
// @Accept json
// @Produce json
// @Param user_book_id path string true "ID progres buku di rak (Bukan master Book ID)"
// @Param request body AddNoteRequest true "Payload isi kutipan dan tag"
// @Success 201 {object} domain.BookNote "Catatan berhasil disimpan"
// @Failure 400 {object} utils.ErrorResponse "Format JSON salah atau Quote kosong"
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan di rak"
// @Failure 500 {object} utils.ErrorResponse "Gagal menyimpan catatan"
// @Router /api/library/{user_book_id}/notes [post]
func (h *BookNoteHandler) AddNote(c *fiber.Ctx) error {
	userBookID := c.Params("user_book_id")

	// Gunakan DTO yang baru dibuat
	var req AddNoteRequest
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

	return c.Status(fiber.StatusCreated).JSON(note)
}

// GetNotes godoc
// @Summary Ambil Daftar Catatan Buku
// @Description Melihat seluruh kutipan dan catatan yang pernah ditulis untuk satu buku spesifik di rak
// @Tags Book Notes
// @Produce json
// @Param user_book_id path string true "ID progres buku di rak"
// @Success 200 {array} domain.BookNote "Daftar catatan"
// @Failure 404 {object} utils.ErrorResponse "Buku tidak ditemukan di rak"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/library/{user_book_id}/notes [get]
func (h *BookNoteHandler) GetNotes(c *fiber.Ctx) error {
	userBookID := c.Params("user_book_id")

	notes, err := h.usecase.GetNotesForBook(c.Context(), userBookID)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	// Best practice: Cegah error null di Frontend jika belum ada notes sama sekali
	if notes == nil {
		notes = make([]*domain.BookNote, 0)
	}

	return c.Status(fiber.StatusOK).JSON(notes)
}
