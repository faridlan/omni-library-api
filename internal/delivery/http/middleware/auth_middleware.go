package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/faridlan/omni-library-api/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.SendError(c, fiber.StatusUnauthorized, "Akses ditolak: Token tidak valid atau kadaluarsa")
		}

		var tokenString string
		parts := strings.Split(authHeader, " ")

		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		} else if len(parts) == 1 {
			tokenString = parts[0]
		} else {
			return utils.SendError(c, fiber.StatusUnauthorized, "Akses ditolak: Format token salah")
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "omnilibrary-super-secret-key"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode enkripsi tidak valid: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return utils.SendError(c, fiber.StatusUnauthorized, "Akses ditolak: Token tidak valid atau kadaluarsa")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return utils.SendError(c, fiber.StatusUnauthorized, "Gagal membaca data token")
		}

		c.Locals("user_id", claims["user_id"].(string))
		c.Locals("role", claims["role"].(string))

		if verified, exists := claims["is_email_verified"]; exists {
			c.Locals("is_email_verified", verified)
		} else {
			// Jika user pakai token lama yang belum ada field is_email_verified-nya
			c.Locals("is_email_verified", false)
		}

		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")

		if role != "admin" {
			return utils.SendError(c, fiber.StatusForbidden, "Akses ditolak: Hanya Admin yang diizinkan melakukan aksi ini")
		}

		return c.Next()
	}
}

func VerifiedOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Langsung ambil status dari Locals yang sudah disiapkan middleware Protected
		isVerified, ok := c.Locals("is_email_verified").(bool)

		// Jika tidak ada, atau nilainya false, tolak akses
		if !ok || !isVerified {
			return utils.SendError(c, fiber.StatusForbidden, "Akses ditolak: Silakan verifikasi email Anda terlebih dahulu")
		}

		// Jika true, lanjutkan ke handler berikutnya
		return c.Next()
	}
}
