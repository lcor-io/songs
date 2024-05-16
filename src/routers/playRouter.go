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
		room := services.Mansion.NewRoomWithId(id, playlist)

		// Create a new player with the session Id
		session := fiber.Locals[string](ctx, "session")
		room.AddPlayer(services.Player{
			Id:      session,
			Guesses: make(map[string]*services.GuessResult),
		})

		return handlers.Render(&ctx, play.Playlist(id))
	})

	playRouter.Post("/:id/guess", func(ctx fiber.Ctx) error {
		session := fiber.Locals[string](ctx, "session")

		room, err := services.Mansion.GetRoom(ctx.Params("id", ""))
		if err != nil {
			return err
		}

		guess := new(Guess)
		if err := ctx.Bind().Form(guess); err != nil {
			return err
		}

		guessResult := room.GuessResult(session, guess.Guess)
		return handlers.Render(&ctx, play.GuessResult(room.PlayedTracks[len(room.PlayedTracks)-1], *guessResult))
	})

	playRouter.Get("/:id/events", func(c fiber.Ctx) error {
		session := fiber.Locals[string](c, "session")

		room, err := services.Mansion.GetRoom(c.Params("id", ""))
		if err != nil {
			return err
		}

		go room.Launch()

		// Create an http stream response
		baseContext := c.Status(fiber.StatusOK).Context()
		baseContext.SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			for {
				select {
				case track := <-room.CurrentTrack:

					htmlWriter := &strings.Builder{}
					components.Audio(track).Render(context.Background(), htmlWriter)
					msg := htmlWriter.String()
					if _, err := fmt.Fprintf(w, "data: %s\n\n", msg); err != nil {
						log.Infof("Error  while flushing: %v. Closing the connection.\n", err)
						room.RemovePlayer(session)
						return
					}

					err := w.Flush()
					if err != nil {
						log.Infof("Error  while flushing: %v. Closing the connection.\n", err)
						room.RemovePlayer(session)
						return
					}

				case <-baseContext.Done():
					log.Info("Client disconnected, closing connection")
					room.RemovePlayer(session)
					return
				}
			}
		}))

		return nil
	}, setSSEHeaders)
}

// Middleware used to set mandatory headers for SSE
func setSSEHeaders(c fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	return c.Next()
}
