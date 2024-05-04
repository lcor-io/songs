package services

// import (
// 	"log"
//
// 	"github.com/google/uuid"
// )

// type Room struct {
// 	// Events are pushed to this channel by the main events-gathering routine
// 	CurrentTrack chan Track
//
// 	// New client connections
// 	NewPlayers chan Player
//
// 	// Closed client connections
// 	ClosedPlayers chan Player
//
// 	// Total client connections
// 	TotalPlayers map[string]Player
// }

// type PlaylistService struct {
// 	ActivePlaylists map[string]*Room
// }
//
// type Player struct {
// 	Id               string
// 	CurrentlyPlaying chan Track
// }
//
// func (stream *Room) listen() {
// 	for {
// 		select {
// 		// Add new available client
// 		case client := <-stream.NewPlayers:
// 			stream.TotalPlayers[client.Id] = client
// 			log.Printf("Client added. %d registered clients", len(stream.TotalPlayers))
//
// 		// Remove closed client
// 		case client := <-stream.ClosedPlayers:
// 			delete(stream.TotalPlayers, client.Id)
// 			log.Printf("Removed client. %d registered clients", len(stream.TotalPlayers))
//
// 		// Broadcast message to client
// 		case eventMsg := <-stream.CurrentTrack:
// 			for PlayerId := range stream.TotalPlayers {
// 				player := stream.TotalPlayers[PlayerId]
// 				player.CurrentlyPlaying <- eventMsg
// 			}
// 		}
// 	}
// }
//
// func (p PlaylistService) RegisterPlaylist(id string) {
// 	if _, ok := p.ActivePlaylists[id]; !ok {
// 		playlistRoom := &Room{
// 			CurrentTrack:  make(chan Track),
// 			NewPlayers:    make(chan Player),
// 			ClosedPlayers: make(chan Player),
// 			TotalPlayers:  make(map[string]Player),
// 		}
// 		p.ActivePlaylists[id] = playlistRoom
// 		log.Println("Create server for playlist ", id)
//
// 		go playlistRoom.listen()
// 	}
// }
//
// func (p PlaylistService) RegisterPlayer(playlistId string) Player {
// 	if stream, ok := p.ActivePlaylists[playlistId]; ok {
// 		// Initialize client channel
// 		newPlayer := Player{
// 			Id:               uuid.NewString(),
// 			CurrentlyPlaying: make(chan Track),
// 		}
//
// 		// Send new connection to event server
// 		stream.NewPlayers <- newPlayer
// 		return newPlayer
// 	}
// 	return Player{}
// }
//
// // func (p PlaylistService) NewRoom(id string) {
// // 	if _, ok := p.ActivePlaylists[id]; !ok {
// // 		//playlist := GetPlaylist(id)
// //
// // 		playlistRoom := &Room{
// // 			CurrentTrack:  make(chan Track),
// // 			NewPlayers:    make(chan chan Track),
// // 			ClosedPlayers: make(chan chan Track),
// // 			TotalPlayers:  make(map[chan Track]bool),
// // 		}
// // 		p.ActivePlaylists[id] = playlistRoom
// // 		log.Println("Create server for playlist ", id)
// //
// // 		// Send track name every 5 seconds, then close the channels
// // 		go func() {
// // 			for _, track := range playlist.Tracks.Items {
// // 				playlistRoom.CurrentTrack <- track.Track
// // 				time.Sleep(time.Second * 29)
// // 			}
// // 			delete(p.ActivePlaylists, id)
// // 		}()
// // 	}
// // }
