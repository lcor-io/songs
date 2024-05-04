package services

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const TRACK_DURATION_SECONDS = 10

type Player struct {
	Id               string
	CurrentlyPlaying chan Track
}

type Room struct {
	Id           string
	Playlist     Playlist
	PlayedTracks map[string]bool
	CurrentTrack chan Track
	Players      map[string]Player
}

func NewRoom(playlist Playlist) *Room {
	return &Room{
		Id:           uuid.New().String(),
		Playlist:     playlist,
		PlayedTracks: make(map[string]bool, len(playlist.Tracks.Items)),
		CurrentTrack: make(chan Track, 1),
		Players:      make(map[string]Player),
	}
}

func NewRoomWithId(id string, playlist Playlist) *Room {
	return &Room{
		Id:       id,
		Playlist: playlist,
	}
}

func (r *Room) Launch() {
	tracks := r.Playlist.Tracks.Items
	for i := 0; i < len(tracks); i++ {
		// Select a track random track from playlist not in already played tracks
		newTrack := tracks[rand.Intn(len(tracks))].Track
		for r.PlayedTracks[newTrack.Name] {
			newTrack = tracks[rand.Intn(len(tracks))].Track
		}

		r.CurrentTrack <- newTrack
		r.PlayedTracks[newTrack.Name] = true
		time.Sleep(TRACK_DURATION_SECONDS * time.Second)
	}
}

func (r *Room) AddPlayer(player Player) {
	id := uuid.New().String()
	r.Players[id] = player
}

func (r *Room) RemovePlayer(player Player) {
	delete(r.Players, player.Id)
}
