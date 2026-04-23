package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDB menginisialisasi koneksi GORM ke PostgreSQL
func InitDB(user, pass, host, port, dbname string) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, pass, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke database: ", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Gagal mendapatkan objek sql.DB: ", err)
	}

	// 3. KONFIGURASI POOLING (Standar Produksi)

	// SetMaxIdleConns: Jumlah pelayan "tetap" yang selalu standby meskipun sepi.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns: Batas MAKSIMAL pelayan. Mencegah database diserang (DDoS) atau overload.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime: Umur maksimal seorang pelayan bekerja sebelum diganti yang baru (refresh koneksi).
	// Sangat penting agar koneksi tidak "basi" atau terputus sepihak oleh firewall/router.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}
