package main

import (
	"log/slog"
	"os"

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
		// Gunakan slog, lalu paksa aplikasi berhenti
		slog.Error("Error loading .env file", slog.String("detail", err.Error()))
		os.Exit(1)
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
	authUsecase := usecase.NewAuthUsecase(userRepo)

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

	// Setup Fiber & Route
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",                                           // Sementara izinkan dari mana saja (Bisa diganti "http://localhost:3000" saat production)
		AllowHeaders: "Origin, Content-Type, Accept, Authorization", // SANGAT PENTING: Izinkan header Authorization!
		AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH",
	}))

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Jakarta", // Sesuaikan dengan zona waktumu
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	// Daftarkan Handler
	myHttp.SetupRoutes(app, authUsecase, bookUsecase, userBookUsecase, bookNoteUsecase)

	// Start Server
	slog.Info("Starting OmniLibrary API Server", slog.String("port", "8080"))
	if err := app.Listen(":8080"); err != nil {
		slog.Error("Server failed to start", slog.String("detail", err.Error()))
		os.Exit(1)
	}
}
