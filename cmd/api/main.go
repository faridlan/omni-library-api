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

	// 2. Inisialisasi Layers (Dependency Injection)
	bookRepo := postgres.NewBookRepository(db)
	bookFetcher := external.NewGoogleBooksFetcher(apiKey)
	bookUsecase := usecase.NewBookUsecase(bookRepo, bookFetcher)

	// 3. Setup Fiber
	app := fiber.New()
	api := app.Group("/api")

	// 4. Register Handler
	http.NewBookHandler(api, bookUsecase)

	// 5. Start Server
	log.Fatal(app.Listen(":8080"))
}
