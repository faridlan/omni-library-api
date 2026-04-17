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

		// 2. Pastikan formatnya "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return utils.SendError(c, fiber.StatusUnauthorized, "Akses ditolak: Format token salah (Gunakan 'Bearer <token>')")
		}

		tokenString := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "omnilibrary-super-secret-key"
		}

		// 3. Verifikasi Keaslian Gelang VIP
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Pastikan algoritma enkripsinya benar (HMAC)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode enkripsi tidak valid: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		// 4. Jika token palsu, kadaluarsa, atau rusak -> USIR!
		if err != nil || !token.Valid {
			return utils.SendError(c, fiber.StatusUnauthorized, "Akses ditolak: Token tidak valid atau kadaluarsa")
		}

		// 5. Jika ASLI, ambil informasi (Claims) dari dalam gelang tersebut
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return utils.SendError(c, fiber.StatusUnauthorized, "Gagal membaca data token")
		}

		// 6. TEMPELKAN ID Warga ke Context (c.Locals)
		// Ini seperti menempelkan name-tag agar Resepsionis (Handler) tahu siapa yang datang
		userID := claims["user_id"].(string)
		c.Locals("user_id", userID)

		// Opsional: Simpan role juga kalau nanti butuh fitur khusus Admin
		c.Locals("role", claims["role"].(string))

		// 7. PERSILAKAN MASUK ke Handler tujuan!
		return c.Next()
	}
}
