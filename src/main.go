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
	"lcor.io/songs/src/utils/middlewares"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmtErr := fmt.Errorf("Error loading .env file: %v", err)
		panic(fmtErr)
	}

	app := fiber.New(fiber.Config{
		AppName: "songs",
	})

	spotify := services.Spotify(os.Getenv("SPOTIFY_CREDENTIALS"))

	// Setup logger in dev
	if os.Getenv("ENV") == "development" {
		app.Use(logger.New(logger.Config{}))
	}

	// Use session middleware
	app.Use(middlewares.SessionMiddleware)

	// Setup compression for incoming requests
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	// Serve static files
	if os.Getenv("ENV") == "development" {
		app.Static("/static", "./static", fiber.Static{
			MaxAge: 0,
		})
	} else {
		app.Static("/static", "./static", fiber.Static{
			MaxAge: 3600,
		})
	}

	// Register routes
	routers.RegisterRoutes(app, spotify)

	log.Fatal(app.Listen(":42068", fiber.ListenConfig{
		EnablePrintRoutes: true,
		EnablePrefork:     false,
	}))
}
