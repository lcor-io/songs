package services

import (
	"errors"
	"math/rand"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3/log"
	"github.com/google/uuid"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"lcor.io/songs/src/utils"
)

const (
	TRACK_DURATION           = 10 * time.Second
	GUESS_VALIDITY_THRESHOLD = 85
)

type Player struct {
	Id string
}

// Contains all active rooms on the server. For now we will keep all rooms in memory
var Mansion = []*Room{}

type Room struct {
	Id           string
	Playlist     Playlist
	PlayedTracks []Track
	CurrentTrack chan Track
	Players      map[string]Player
}

func NewRoom(playlist Playlist) *Room {
	return NewRoomWithId(uuid.New().String(), playlist)
}

func NewRoomWithId(id string, playlist Playlist) *Room {
	newRoom := &Room{
		Id:           id,
		Playlist:     playlist,
		PlayedTracks: make([]Track, 0, len(playlist.Tracks.Items)),
		CurrentTrack: make(chan Track),
		Players:      make(map[string]Player),
	}
	Mansion = append(Mansion, newRoom)
	return newRoom
}

func GetRoomById(id string) (*Room, error) {
	for _, room := range Mansion {
		if room.Id == id {
			return room, nil
		}
	}
	return nil, errors.New("Room not found")
}

func (r *Room) Launch() {
	defer close(r.CurrentTrack)

	playlistTracks := r.Playlist.Tracks.Items

	for i := 0; i < len(playlistTracks); i++ {
		// Select a track random track from playlist not in already played tracks
		newTrack := playlistTracks[rand.Intn(len(playlistTracks))].Track
		newTrackAlreadyPlayed := slices.ContainsFunc(r.PlayedTracks, func(t Track) bool {
			return t.Name == newTrack.Name
		})
		for newTrackAlreadyPlayed {
			newTrack = playlistTracks[rand.Intn(len(playlistTracks))].Track
		}

		r.CurrentTrack <- newTrack
		r.PlayedTracks = append(r.PlayedTracks, newTrack)
		time.Sleep(TRACK_DURATION)
	}
}

func (r *Room) GuessResult(guess string) bool {
	currentTrack := r.PlayedTracks[len(r.PlayedTracks)-1]

	normalizedGuess := utils.Normalize(guess)
	normalizedTitle := utils.Normalize(currentTrack.Name)

	guessLen := utf8.RuneCountInString(normalizedGuess)
	titleLen := utf8.RuneCountInString(normalizedTitle)
	score := fuzzy.LevenshteinDistance(normalizedGuess, normalizedTitle)

	normalizedScore := 100 * (float32(max(guessLen, titleLen)) - float32(score)) / float32(max(guessLen, titleLen))

	return normalizedScore >= GUESS_VALIDITY_THRESHOLD
}

func (r *Room) AddPlayer(player Player) {
	log.Infof("Player %s joined room %s", player.Id, r.Id)

	id := uuid.New().String()
	r.Players[id] = player
}

func (r *Room) RemovePlayer(id string) {
	log.Infof("Player %s leaved room %s", id, r.Id)

	delete(r.Players, id)

	// The room is empty, remove it
	if len(r.Players) == 0 {
		log.Infof("Room %s is empty, removing it", r.Id)
		close(r.CurrentTrack)
		Mansion = slices.DeleteFunc(Mansion, func(room *Room) bool {
			return room.Id == r.Id
		})
		r = nil
	}
}
