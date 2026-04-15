package main

import (
	"log"
	"os"

	"github.com/faridlan/omni-library-api/internal/delivery/http"
	"github.com/faridlan/omni-library-api/internal/repository/external"
	"github.com/faridlan/omni-library-api/internal/repository/postgres"
	"github.com/faridlan/omni-library-api/internal/usecase"
	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v2"
)

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

	db := postgres.InitDB(dbUser, dbPassword, dbHost, dbPort, dbName)

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
	api := app.Group("/api")

	// Daftarkan Handler
	http.NewBookHandler(api, bookUsecase)
	http.NewUserBookHandler(api, userBookUsecase)
	http.NewBookNoteHandler(api, bookNoteUsecase)

	// Start Server
	log.Fatal(app.Listen(":8080"))
}
