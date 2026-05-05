package http

import (
	"github.com/faridlan/omni-library-api/internal/delivery/http/dto"
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// Resepsionis Khusus Auth
type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(uc domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: uc,
	}
}

// ==========================================
// HELPER: MAPPING ENTITY KE RESPONSE DTO
// ==========================================
func toUserResponse(user *domain.User) dto.UserResponse {
	if user == nil {
		return dto.UserResponse{}
	}
	return dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
		// Format tanggal bisa dibiarkan sebagai string sesuai dengan DTO yang kamu buat.
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Register godoc
// @Summary      Registrasi User Baru
// @Description  Mendaftarkan pengguna baru ke dalam sistem dan melakukan hashing pada password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Payload registrasi pengguna baru"
// @Success      201 {object} utils.SuccessResponse[dto.UserResponse] "User berhasil dibuat"
// @Failure      400 {object} utils.ErrorResponse "Format JSON salah atau validasi gagal"
// @Failure      409 {object} utils.ErrorResponse "Email sudah terdaftar (Conflict)"
// @Failure      500 {object} utils.ErrorResponse "Internal Server Error"
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON tidak valid")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	reqInput := domain.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := h.authUsecase.Register(c.Context(), reqInput)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	// Menggunakan Helper Mapper
	res := toUserResponse(user)

	return utils.SendSuccess(c, fiber.StatusCreated, "User berhasil dibuat", res)
}

// Login godoc
// @Summary      Login User (Dapatkan JWT)
// @Description  Autentikasi email dan password pengguna untuk mendapatkan Token JWT.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Payload kredensial login"
// @Success      200 {object} utils.SuccessResponse[dto.TokenResponse] "Login berhasil, berisi access_token dan refresh_token"
// @Failure      400 {object} utils.ErrorResponse "Format JSON salah atau validasi gagal"
// @Failure      401 {object} utils.ErrorResponse "Email atau Password salah"
// @Failure      404 {object} utils.ErrorResponse "Email tidak ditemukan"
// @Failure      500 {object} utils.ErrorResponse "Internal Server Error"
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON tidak valid")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	reqInput := domain.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	accessToken, refreshToken, err := h.authUsecase.Login(c.Context(), reqInput)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	res := dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Login berhasil", res)
}

// Refresh godoc
// @Summary      Perbarui Access Token
// @Description  Menukarkan Refresh Token lama (berumur 7 hari) dengan Access Token baru (15 menit). Cocok dipanggil diam-diam oleh Frontend saat mendapat error 401.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshRequest true "Payload Refresh Token"
// @Success      200 {object} utils.SuccessResponse[dto.TokenResponse] "Berhasil mendapat access token baru"
// @Failure      400 {object} utils.ErrorResponse "Format salah atau token tidak valid"
// @Failure      401 {object} utils.ErrorResponse "Token expired atau ditolak"
// @Router       /api/auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req dto.RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	newAccessToken, err := h.authUsecase.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	resdata := dto.TokenResponse{
		AccessToken: newAccessToken,
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Access Token berhasil diperbarui", resdata)
}
