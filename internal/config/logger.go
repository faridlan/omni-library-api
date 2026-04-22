package config

import (
	"log/slog"
	"os"
)

// InitLogger mengatur slog untuk mencetak log dalam format JSON
func InitLogger() {
	// Kita menggunakan JSONHandler agar outputnya berformat JSON
	// HandlerOptions memungkinkan kita mengatur level log (Info, Warn, Error, Debug)
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	// Jadikan logger ini sebagai default global
	// Jadi kita cukup memanggil slog.Info() atau slog.Error() di mana saja
	slog.SetDefault(logger)
}
