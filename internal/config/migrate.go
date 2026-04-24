package config // Sesuaikan dengan nama package-mu

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// GANTI INI dengan module path aslimu yang mengarah ke folder migrations
	"github.com/faridlan/omni-library-api/db/migrations"
)

func RunDBMigration(dbURL string) {
	// Membaca file SQL dari dalam binary Go menggunakan iofs
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("Gagal membaca file migrasi embed: %v", err)
	}

	// Membuat instance migrasi baru
	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		log.Fatalf("Gagal membuat instance migrasi: %v", err)
	}

	// Mengeksekusi migrasi UP
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Gagal menjalankan migrasi: %v", err)
	}

	log.Println("✅ Database migrasi berhasil dijalankan!")
}
