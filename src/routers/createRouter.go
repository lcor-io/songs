package routers

import (
	"time"

	"github.com/gofiber/fiber/v3"

	playlist "lcor.io/songs/src/components/playlist"
	"lcor.io/songs/src/models"
	"lcor.io/songs/src/pages/create"
	"lcor.io/songs/src/repositories"
	"lcor.io/songs/src/services"
	"lcor.io/songs/src/utils"
)

func RegisterCreateRoutes(router fiber.Router, spotify *services.SpotifyService, repo *repositories.RoomRepository) {
	router.Get("/", func(c fiber.Ctx) error {
		return utils.TemplRender(&c, pages.Create())
	})

	router.Get("/featured", func(ctx fiber.Ctx) error {
		spotifyPlaylists := spotify.GetFeaturedPlaylist()
		playlists := make([]models.Playlist, 0, len(spotifyPlaylists.Playlists.Items))
		for _, p := range spotifyPlaylists.Playlists.Items {
			playlists = append(playlists, p.ToPlaylist())
		}

		ctx.Set("Cache-Control", "max-age=60, stale-while-revalidate=3600")
		return utils.TemplRender(&ctx, playlist.InlinePlaylists("Featured Playlists", playlists))
	})

	router.Post("/:id", func(c fiber.Ctx) error {
		id := c.Params("id")

		playlist := spotify.GetPlaylist(id)
		room := services.Mansion.NewRoom(playlist)

		c.Set("HX-Location", "/play/"+room.Id)
		return c.SendStatus(fiber.StatusCreated)
	})
}
