package routers

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/valyala/fasthttp"

	"lcor.io/songs/src/components"
	"lcor.io/songs/src/handlers"
	"lcor.io/songs/src/services"

	"lcor.io/songs/src/pages/play"
)

type Guess struct {
	Guess string `form:"guess"`
}

func RegisterPlayRoutes(app *fiber.App, spotify *services.SpotifyService) {
	playRouter := app.Group("/play")

	playRouter.Get("/", func(ctx fiber.Ctx) error {
		return handlers.Render(&ctx, play.Play())
	})

	playRouter.Get("/featured", func(ctx fiber.Ctx) error {
		playlists := spotify.GetFeaturedPlaylist()
		ctx.Set("Cache-Control", "max-age=60, stale-while-revalidate=3600")
		return handlers.Render(&ctx, components.InlinePlaylists("Featured Playlists", playlists))
	})

	playRouter.Get("/:id", func(ctx fiber.Ctx) error {
		id := ctx.Params("id")

		// Create room with the specified playlist
		playlist := spotify.GetPlaylist(id)
		room := services.NewRoomWithId(id, playlist)

		// Create a new player with the session Id
		session := ctx.Locals("session").(string)
		room.AddPlayer(services.Player{
			Id: session,
		})

		return handlers.Render(&ctx, play.Playlist(id))
	})

	playRouter.Post("/:id/guess", func(ctx fiber.Ctx) error {
		id := ctx.Params("id")

		guess := new(Guess)

		if err := ctx.Bind().Form(guess); err != nil {
			return err
		}

		room, err := services.GetRoomById(id)
		if err != nil {
			return err
		}
		if room.GuessResult(guess.Guess) {
			return ctx.SendString(fmt.Sprintf("%s is a correct guess!", guess.Guess))
		}

		return ctx.SendString(fmt.Sprintf("%s is an incorrect guess!", guess.Guess))
	})

	playRouter.Get("/:id/events", func(c fiber.Ctx) error {
		id := c.Params("id")
		session := c.Locals("session").(string)

		room, err := services.GetRoomById(id)
		if err != nil {
			return err
		}

		// Set headers
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		go room.Launch()

		// Create an http stream response
		c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			for track := range room.CurrentTrack {

				// Create the HTML related to the event
				htmlWriter := &strings.Builder{}
				components.Audio(track).Render(context.Background(), htmlWriter)
				msg := htmlWriter.String()
				fmt.Fprintf(w, "data: %s\n\n", msg)

				// Send the event to the client
				err := w.Flush()
				log.Info("Sending event to client")
				if err != nil {
					log.Infof("Error  while flushing: %v. Closing the connection.\n", err)
					room.RemovePlayer(session)
					break
				}
			}
		}))

		return nil
	})
}
