package routers

import (
	"github.com/gofiber/fiber/v3"

	"lcor.io/songs/src/pages"
	"lcor.io/songs/src/services"
	"lcor.io/songs/src/utils"
)

func RegisterRoutes(app *fiber.App, spotify *services.SpotifyService) {
	app.Get("/", func(c fiber.Ctx) error {
		return utils.TemplRender(&c, pages.Landing())
	})

	RegisterCreateRoutes(app.Group("/create"), spotify)
	RegisterPlayRoutes(app.Group("/play"))
}
