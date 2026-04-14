package main

import (
	"log"

	"github.com/faridlan/omni-library-api/internal/delivery/http"
	"github.com/faridlan/omni-library-api/internal/repository/external"
	"github.com/faridlan/omni-library-api/internal/repository/postgres"
	"github.com/faridlan/omni-library-api/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Inisialisasi Database
	// Sesuaikan dengan kredensial Makefile/Docker kamu
	db := postgres.InitDB("nullhakim", "NullHakimNostra123", "localhost", "5432", "omnilibrary")

	// 2. Inisialisasi Layers (Dependency Injection)
	bookRepo := postgres.NewBookRepository(db)
	bookFetcher := external.NewGoogleBooksFetcher()
	bookUsecase := usecase.NewBookUsecase(bookRepo, bookFetcher)

	// 3. Setup Fiber
	app := fiber.New()
	api := app.Group("/api")

	// 4. Register Handler
	http.NewBookHandler(api, bookUsecase)

	// 5. Start Server
	log.Fatal(app.Listen(":8080"))
}
