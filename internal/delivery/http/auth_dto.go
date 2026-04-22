package http

// DTO untuk validasi Input
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"` // Minimal 6 karakter agar aman
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// DTO untuk Response Registrasi (Tanpa Password!)
type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"` // Format string agar mudah dibaca di JSON
}

// DTO untuk Response Login (Mengembalikan Token)
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Struct untuk menangkap payload JSON di endpoint /refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
