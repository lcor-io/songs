package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/logger"

	"github.com/joho/godotenv"

	"lcor.io/songs/src/routers"
	"lcor.io/songs/src/services"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmtErr := fmt.Errorf("Error loading .env file: %v", err)
		panic(fmtErr)
	}

	spotify := services.Spotify(os.Getenv("SPOTIFY_CREDENTIALS"))

	app := fiber.New(fiber.Config{
		AppName: "songs",
	})

	// Setup logger in dev
	if os.Getenv("ENV") == "development" {
		app.Use(logger.New(logger.Config{}))
	}

	// Setup compression for incoming requests
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	// Serve static files
	app.Static("/static/css", "./static/css", fiber.Static{
		MaxAge: 3600,
	})

	// Register routes
	routers.RegisterRoutes(app, spotify)

	log.Fatal(app.Listen(":42069", fiber.ListenConfig{
		EnablePrintRoutes: true,
	}))
}
