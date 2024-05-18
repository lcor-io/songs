package routers

import (
	"github.com/gofiber/fiber/v3"

	"lcor.io/songs/src/utils"
	"lcor.io/songs/src/pages"
	"lcor.io/songs/src/services"
)

func RegisterRoutes(app *fiber.App, spotify *services.SpotifyService) {
	app.Get("/", func(c fiber.Ctx) error {
		return utils.TemplRender(&c, pages.Landing())
	})
	RegisterPlayRoutes(app, spotify)
}
