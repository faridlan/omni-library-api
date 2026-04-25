package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/faridlan/omni-library-api/docs"
	"github.com/faridlan/omni-library-api/internal/config"
	myHttp "github.com/faridlan/omni-library-api/internal/delivery/http"
	"github.com/faridlan/omni-library-api/internal/repository/external"
	"github.com/faridlan/omni-library-api/internal/repository/postgres"
	"github.com/faridlan/omni-library-api/internal/usecase"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	_ "github.com/faridlan/omni-library-api/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
)

// @title OmniLibrary API
// @version 1.0
// @description Ini adalah dokumentasi API untuk MVP OmniLibrary.
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token dengan format: Bearer {token}
func main() {

	err := godotenv.Load()
	if err != nil {
		slog.Warn("File .env tidak ditemukan, menggunakan environment variable dari sistem")
	}

	// Ambil nilai dengan os.Getenv
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	apiKey := os.Getenv("GOOGLE_BOOKS_API_KEY")

	db := config.InitDB(dbUser, dbPassword, dbHost, dbPort, dbName)

	//Fitur Authentication & Authorization
	userRepo := postgres.NewUserRepository(db)
	authRepo := postgres.NewAuthRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepo, authRepo)

	// Fitur Book Metadata
	bookRepo := postgres.NewBookRepository(db)
	bookFetcher := external.NewGoogleBooksFetcher(apiKey)
	bookUsecase := usecase.NewBookUsecase(bookRepo, bookFetcher)

	// Fitur Reading Tracker
	userBookRepo := postgres.NewUserBookRepository(db)
	userBookUsecase := usecase.NewUserBookUsecase(userBookRepo, bookRepo)

	// Fitur Book Notes (Quotes & Tags)
	bookNoteRepo := postgres.NewBookNoteRepository(db)
	bookNoteUsecase := usecase.NewBookNoteUsecase(bookNoteRepo, userBookRepo)

	dbURL := os.Getenv("DB_URL")

	// 1. Jalankan Migrasi Otomatis!
	config.RunDBMigration(dbURL)

	swaggerHost := os.Getenv("SWAGGER_HOST")
	if swaggerHost != "" {
		docs.SwaggerInfo.Host = swaggerHost
	}

	// Setup Fiber & Route
	app := fiber.New()

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "*" // Fallback untuk kemudahan di lokal
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     frontendURL,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: false,
	}))

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Jakarta", // Sesuaikan dengan zona waktumu
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	// Daftarkan Handler
	myHttp.SetupRoutes(app, authUsecase, bookUsecase, userBookUsecase, bookNoteUsecase)

	// Endpoint Tracer Bullet untuk mengetes CI/CD
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Hello dari Staging! CI/CD Otomatis berhasil mendarat dengan mulus. 🚀",
			"version": "1.0.1-beta",
		})
	})

	// Start Server
	go func() {
		slog.Info("Starting OmniLibrary API Server", slog.String("port", "8080"))
		if err := app.Listen(":8080"); err != nil {
			slog.Error("Server failed to start", slog.String("detail", err.Error()))
		}
	}()

	// 2. Buat "Jebakan" Sinyal (Menunggu CTRL+C atau instruksi Docker Stop)
	quit := make(chan os.Signal, 1)
	// SIGINT = CTRL+C, SIGTERM = Sinyal kill dari Docker/Linux
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit // Main Thread akan BERHENTI di sini menunggu sinyal masuk ke channel

	// 3. Menutup Warung dengan Sopan (Graceful Shutdown)
	slog.Info("Menerima sinyal mati, mematikan server dengan sopan...")

	// Fiber akan menolak request baru, tapi menunggu request yang sedang berjalan selesai
	if err := app.Shutdown(); err != nil {
		slog.Error("Server dipaksa mati karena error", slog.String("detail", err.Error()))
	}

	slog.Info("OmniLibrary API Server berhasil dimatikan dengan aman. Sampai jumpa!")
}
