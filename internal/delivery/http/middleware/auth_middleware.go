package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected adalah Satpam kita!
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Cek Header "Authorization"
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.SendError(c, fiber.StatusUnauthorized, "Akses ditolak: Token tidak ditemukan")
		}

		// 2. LOGIKA SATPAM YANG AMAN DARI PANIC (Index Out Of Range)
		var tokenString string
		parts := strings.Split(authHeader, " ")

		if len(parts) == 2 && parts[0] == "Bearer" {
			// Jika formatnya "Bearer eyJhb..."
			tokenString = parts[1]
		} else if len(parts) == 1 {
			// Jika formatnya langsung "eyJhb..." (Tanpa Bearer)
			tokenString = parts[0]
		} else {
			return utils.SendError(c, fiber.StatusUnauthorized, "Akses ditolak: Format token salah")
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "omnilibrary-super-secret-key"
		}

		// 3. Verifikasi Keaslian Gelang VIP
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode enkripsi tidak valid: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		// 4. Usir jika token palsu/kadaluarsa
		if err != nil || !token.Valid {
			return utils.SendError(c, fiber.StatusUnauthorized, "Akses ditolak: Token tidak valid atau kadaluarsa")
		}

		// 5. Tempelkan ID ke Context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return utils.SendError(c, fiber.StatusUnauthorized, "Gagal membaca data token")
		}

		c.Locals("user_id", claims["user_id"].(string))
		c.Locals("role", claims["role"].(string))

		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil role yang tadi ditempelkan oleh middleware Protected()
		role := c.Locals("role")

		// Jika bukan admin, usir dengan 403 Forbidden
		if role != "admin" {
			return utils.SendError(c, fiber.StatusForbidden, "Akses ditolak: Hanya Admin yang diizinkan melakukan aksi ini")
		}

		// Jika admin, persilakan lewat
		return c.Next()
	}
}
