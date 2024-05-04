package routers

import (
	"github.com/gofiber/fiber/v3"

	"lcor.io/songs/src/handlers"
	"lcor.io/songs/src/pages"
	"lcor.io/songs/src/services"
)

func RegisterRoutes(app *fiber.App, spotify *services.SpotifyService) {
	app.Get("/", func(c fiber.Ctx) error {
		return handlers.Render(&c, pages.Landing())
	})
	RegisterPlayRoutes(app, spotify)
}
