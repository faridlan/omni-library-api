package http

import (
	"github.com/faridlan/omni-library-api/internal/delivery/http/dto"
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userUsecase domain.UserUsecase
}

func NewUserHandler(router fiber.Router, userUsecase domain.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
	}
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Mengambil data profil pengguna yang sedang login berdasarkan token JWT
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success 200 {object} utils.SuccessResponse[dto.UserProfileResponse] "Berhasil mengambil profil"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Token tidak ada/salah)"
// @Failure 404 {object} utils.ErrorResponse "User tidak ditemukan"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router       /api/users/me [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	// Ambil userID dari JWT Middleware
	userID := c.Locals("user_id").(string)

	user, err := h.userUsecase.GetProfile(c.Context(), userID)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	res := dto.UserProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Berhasil mengambil profil", res)
}

// UpdateProfile godoc
// @Summary      Update current user profile
// @Description  Memperbarui data profil (seperti nama) pengguna yang sedang login
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.UpdateProfileRequest true "Data profil yang ingin diupdate"
// @Success      200  {object}  utils.SuccessResponse[dto.UserProfileResponse] "Profil berhasil diperbarui"
// @Failure      400  {object}  utils.ErrorResponse "Bad Request (Validasi gagal)"
// @Failure      401  {object}  utils.ErrorResponse "Unauthorized"
// @Failure      404  {object}  utils.ErrorResponse "User tidak ditemukan"
// @Failure      500  {object}  utils.ErrorResponse "Internal Server Error"
// @Router       /api/users/me [put]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format input tidak valid")
	}

	user, err := h.userUsecase.UpdateProfile(c.Context(), userID, &req)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	res := dto.UserProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Profil berhasil diperbarui", res)
}

// UpdatePassword godoc
// @Summary      Update user password
// @Description  Memperbarui kata sandi pengguna yang sedang login (membutuhkan kata sandi lama)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.UpdatePasswordRequest true "Data kata sandi lama dan baru"
// @Success      200  {object}  utils.SuccessResponse[dto.UserProfileResponse] "Password berhasil diperbarui"
// @Failure      400  {object}  utils.ErrorResponse "Bad Request (Validasi gagal / Password lama salah)"
// @Failure      401  {object}  utils.ErrorResponse "Unauthorized"
// @Failure      404  {object}  utils.ErrorResponse "User tidak ditemukan"
// @Failure      500  {object}  utils.ErrorResponse "Internal Server Error"
// @Security     BearerAuth
// @Router       /api/users/me/password [put]
func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format input tidak valid")
	}

	err := h.userUsecase.UpdatePassword(c.Context(), userID, &req)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Password berhasil diperbarui", nil)
}
