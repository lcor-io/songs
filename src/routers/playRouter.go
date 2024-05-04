package routers

import (
	// "context"
	// "fmt"
	// "io"
	// "log"
	// "net/http"
	// "strings"
	// "time"

	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"

	"lcor.io/songs/src/components"
	"lcor.io/songs/src/handlers"
	"lcor.io/songs/src/services"

	"lcor.io/songs/src/pages/play"
)

// var playlistChannels = make(map[string]*Event)
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
		return handlers.Render(&ctx, play.Playlist(id))
	})

	playRouter.Get("/:id/events", func(c fiber.Ctx) error {
		id := c.Params("id")

		// Set headers
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		// Create room with the specified playlist
		playlist := spotify.GetPlaylist(id)
		room := services.NewRoom(playlist)
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
				if err != nil {
					fmt.Printf("Error  while flushing: %v. Closing the connection.\n", err)
					break
				}
			}
		}))

		return nil
	})
}

// 	playRouter.GET("/:id", func(ctx *gin.Context) {
// 		id := ctx.Param("id")
//
// 		// Create playlist server if needed
// 		if _, ok := playlistChannels[id]; !ok {
// 			playlist := spotify.GetPlaylist(id)
//
// 			playlistServer := NewServer()
// 			playlistChannels[id] = playlistServer
// 			log.Println("Create server for playlist ", id)
//
// 			// Send track name every 5 seconds, then close the channels
// 			go func() {
// 				for _, track := range playlist.Tracks.Items {
// 					playlistServer.Message <- track.Track
// 					time.Sleep(time.Second * 29)
// 				}
// 				// for clientMessageChan := range playlistServer.TotalClients {
// 				// 	playlistServer.ClosedClients <- clientMessageChan
// 				// }
// 				delete(playlistChannels, id)
// 			}()
// 		}
//
// 		ctx.HTML(http.StatusOK, "", play.Playlist(id))
// 	})
//
// 	playRouter.GET("/:id/events", HeadersMiddleware(), serveHTTP(), func(ctx *gin.Context) {
// 		id := ctx.Param("id")
// 		v, ok := ctx.Get(id)
// 		if !ok {
// 			return
// 		}
// 		clientChan, ok := v.(ClientChan)
// 		if !ok {
// 			return
// 		}
//
// 		ctx.Stream(func(w io.Writer) bool {
// 			if msg, ok := <-clientChan; ok {
// 				htmlWriter := &strings.Builder{}
// 				components.Audio(msg).Render(context.Background(), htmlWriter)
// 				ctx.SSEvent("message", htmlWriter.String())
// 				return true
// 			}
// 			return false
// 		})
// 	})
// }
//
// // It keeps a list of clients those are currently attached
// // and broadcasting events to those clients.
// type Event struct {
// 	// Events are pushed to this channel by the main events-gathering routine
// 	Message chan services.Track
//
// 	// New client connections
// 	NewClients chan chan services.Track
//
// 	// Closed client connections
// 	ClosedClients chan chan services.Track
//
// 	// Total client connections
// 	TotalClients map[chan services.Track]bool
// }
//
// // New event messages are broadcast to all registered client connection channels
// type ClientChan chan services.Track
//
// // Initialize event and Start processing requests
// func NewServer() (event *Event) {
// 	event = &Event{
// 		Message:       make(chan services.Track),
// 		NewClients:    make(chan chan services.Track),
// 		ClosedClients: make(chan chan services.Track),
// 		TotalClients:  make(map[chan services.Track]bool),
// 	}
//
// 	go event.listen()
// 	return
// }
//
// // It Listens all incoming requests from clients.
// // Handles addition and removal of clients and broadcast messages to clients.
// func (stream *Event) listen() {
// 	for {
// 		select {
// 		// Add new available client
// 		case client := <-stream.NewClients:
// 			stream.TotalClients[client] = true
// 			log.Printf("Client added. %d registered clients", len(stream.TotalClients))
//
// 		// Remove closed client
// 		case client := <-stream.ClosedClients:
// 			delete(stream.TotalClients, client)
// 			close(client)
// 			log.Printf("Removed client. %d registered clients", len(stream.TotalClients))
//
// 		// Broadcast message to client
// 		case eventMsg := <-stream.Message:
// 			for clientMessageChan := range stream.TotalClients {
// 				clientMessageChan <- eventMsg
// 			}
// 		}
// 	}
// }
//
// func serveHTTP() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		playlistId := c.Param("id")
//
// 		// Get the stream
// 		stream, ok := playlistChannels[playlistId]
// 		if !ok {
// 			fmt.Println("No stream found")
// 			return
// 		}
//
// 		// Initialize client channel
// 		clientChan := make(ClientChan)
//
// 		// Send new connection to event server
// 		stream.NewClients <- clientChan
//
// 		defer func() {
// 			log.Println("Closing client channel")
// 			// Send closed connection to event server
// 			stream.ClosedClients <- clientChan
// 		}()
//
// 		// go func() {
// 		// 	<-c.Done()
// 		// 	log.Println("Closing client channel")
// 		// 	stream.ClosedClients <- clientChan
// 		// }()
//
// 		c.Set(playlistId, clientChan)
//
// 		c.Next()
// 	}
// }
//
// func HeadersMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Writer.Header().Set("Content-Type", "text/event-stream")
// 		c.Writer.Header().Set("Cache-Control", "no-cache")
// 		c.Writer.Header().Set("Connection", "keep-alive")
// 		c.Writer.Header().Set("Transfer-Encoding", "chunked")
// 		c.Next()
// 	}
// }
