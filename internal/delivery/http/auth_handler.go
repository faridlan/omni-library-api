package http

import (
	"github.com/faridlan/omni-library-api/internal/domain"
	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// Resepsionis Khusus Auth
type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(router fiber.Router, uc domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: uc,
	}

	// Buat sub-grup khusus auth
	// authGroup := router.Group("/auth")
	// authGroup.Post("/register", handler.Register)
	// authGroup.Post("/login", handler.Login)
}

// Register godoc
// @Summary Registrasi User Baru
// @Description Mendaftarkan pengguna baru ke dalam sistem dan melakukan hashing pada password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Payload registrasi pengguna baru"
// @Success 201 {object} UserResponse "Registrasi berhasil (mengembalikan data user tanpa password)"
// @Failure 400 {object} utils.ErrorResponse "Format JSON salah atau validasi gagal"
// @Failure 409 {object} utils.ErrorResponse "Email sudah terdaftar (Conflict)"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// 1. Tangkap JSON Body
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON tidak valid")
	}

	// 2. Validasi Input (Memakai utils validator yang sudah kita buat)
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	// 3. Lempar ke Otak (Usecase)
	user, err := h.authUsecase.Register(c.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	// 4. Kembalikan Response Sukses (Jangan pernah kembalikan password di response!)
	res := UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"), // Format waktu agar cantik
	}

	// Kembalikan DTO tersebut
	return c.Status(fiber.StatusCreated).JSON(res)
}

// Login godoc
// @Summary Login User (Dapatkan JWT)
// @Description Autentikasi email dan password pengguna untuk mendapatkan Token JWT.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Payload kredensial login"
// @Success 200 {object} TokenResponse "Login berhasil (mengembalikan token JWT)"
// @Failure 400 {object} utils.ErrorResponse "Format JSON salah atau validasi gagal"
// @Failure 401 {object} utils.ErrorResponse "Email atau Password salah"
// @Failure 404 {object} utils.ErrorResponse "Email tidak ditemukan"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	// 1. Tangkap JSON Body
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON tidak valid")
	}

	// 2. Validasi Input
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	// 3. Lempar ke Otak (Usecase) untuk divalidasi dan dibuatkan JWT
	accessToken, refreshToken, err := h.authUsecase.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return utils.HandleDomainError(c, err)
	}

	// Buat object DTO Response
	res := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// Kembalikan DTO tersebut
	return c.Status(fiber.StatusOK).JSON(res)
}

// Refresh godoc
// @Summary Perbarui Access Token
// @Description Menukarkan Refresh Token lama (berumur 7 hari) dengan Access Token baru (15 menit). Cocok dipanggil diam-diam oleh Frontend saat mendapat error 401.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Payload Refresh Token"
// @Success 200 {object} map[string]string "Berhasil mendapat access token baru"
// @Failure 400 {object} utils.ErrorResponse "Format salah atau token tidak valid"
// @Failure 401 {object} utils.ErrorResponse "Token expired atau ditolak"
// @Router /api/auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest

	// 1. Tangkap JSON dari Frontend
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON salah")
	}

	// 2. Validasi apakah refresh_token kosong
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	// 3. Serahkan token lama ke Usecase untuk ditukarkan
	newAccessToken, err := h.authUsecase.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		// Jika token palsu, expired, atau tidak ada di DB, kita usir
		return utils.HandleDomainError(c, err)
	}

	// 4. Berikan Access Token baru ke Frontend
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Access Token berhasil diperbarui",
		"access_token": newAccessToken,
	})
}
