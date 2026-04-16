package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Gunakan satu instance validator untuk seluruh aplikasi (Singleton)
var validate = validator.New()

// ValidateStruct memeriksa struct dan mengembalikan pesan error yang ramah dibaca
func ValidateStruct(data any) error {
	err := validate.Struct(data)
	if err != nil {
		// Jika errornya berasal dari validasi, kita percantik pesannya
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Kita ambil error dari field pertama yang gagal saja agar rapi
			firstErr := validationErrors[0]
			return fmt.Errorf("kolom '%s' gagal pada validasi '%s'", firstErr.Field(), firstErr.Tag())
		}
		return err
	}
	return nil
}
