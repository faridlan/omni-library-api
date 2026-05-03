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

	config.RunDBMigration(dbURL)

	swaggerHost := os.Getenv("SWAGGER_HOST")
	if swaggerHost != "" {
		docs.SwaggerInfo.Host = swaggerHost
	}

	app := fiber.New()

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "*"
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
		TimeZone:   "Asia/Jakarta",
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	// Daftarkan Handler
	myHttp.SetupRoutes(app, authUsecase, bookUsecase, userBookUsecase, bookNoteUsecase)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		slog.Info("Starting OmniLibrary API Server", slog.String("port", port))
		if err := app.Listen(":" + port); err != nil {
			slog.Error("Server failed to start", slog.String("detail", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	slog.Info("Menerima sinyal mati, mematikan server dengan sopan...")

	if err := app.Shutdown(); err != nil {
		slog.Error("Server dipaksa mati karena error", slog.String("detail", err.Error()))
	}

	slog.Info("OmniLibrary API Server berhasil dimatikan dengan aman. Sampai jumpa!")
}
