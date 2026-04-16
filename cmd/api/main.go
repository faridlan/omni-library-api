package main

import (
	"log"
	"os"

	"github.com/faridlan/omni-library-api/internal/config"
	myHttp "github.com/faridlan/omni-library-api/internal/delivery/http"
	"github.com/faridlan/omni-library-api/internal/repository/external"
	"github.com/faridlan/omni-library-api/internal/repository/postgres"
	"github.com/faridlan/omni-library-api/internal/usecase"
	"github.com/joho/godotenv"

	_ "github.com/faridlan/omni-library-api/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// @title OmniLibrary API
// @version 1.0
// @description Ini adalah dokumentasi API untuk MVP OmniLibrary.
// @host localhost:8080
// @BasePath /
func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Ambil nilai dengan os.Getenv
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	apiKey := os.Getenv("GOOGLE_BOOKS_API_KEY")

	db := config.InitDB(dbUser, dbPassword, dbHost, dbPort, dbName)

	// Fitur Book Metadata
	bookRepo := postgres.NewBookRepository(db)
	bookFetcher := external.NewGoogleBooksFetcher(apiKey)
	bookUsecase := usecase.NewBookUsecase(bookRepo, bookFetcher)

	// Fitur Reading Tracker
	userBookRepo := postgres.NewUserBookRepository(db)
	userBookUsecase := usecase.NewUserBookUsecase(userBookRepo)

	// Fitur Book Notes (Quotes & Tags)
	bookNoteRepo := postgres.NewBookNoteRepository(db)
	bookNoteUsecase := usecase.NewBookNoteUsecase(bookNoteRepo)

	// Setup Fiber & Route
	app := fiber.New()
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Daftarkan Handler
	myHttp.SetupRoutes(app, bookUsecase, userBookUsecase, bookNoteUsecase)

	// Start Server
	log.Fatal(app.Listen(":8080"))
}
