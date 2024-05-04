package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/google/uuid"

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

	app := fiber.New(fiber.Config{
		AppName: "songs",
	})

	spotify := services.Spotify(os.Getenv("SPOTIFY_CREDENTIALS"))

	// Setup logger in dev
	if os.Getenv("ENV") == "development" {
		app.Use(logger.New(logger.Config{}))
	}

	// Use session middleware
	store := session.New(session.Config{
		Expiration:     time.Hour * 24 * 30,
		KeyLookup:      "cookie:songs_session",
		KeyGenerator:   uuid.New().String,
		CookieSecure:   true,
		CookieHTTPOnly: true,
	})
	app.Use(func(c fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting session")
		}
		fmt.Printf("Session ID: %s\n", sess.ID())
		if err := sess.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error setting session")
		}
		return c.Next()
	})

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
