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

type GuessResult struct {
	Title   bool
	Artists map[string]bool
}

type Player struct {
	Id      string
	Guesses map[string]*GuessResult
}

// Contains all active rooms on the server. For now we will keep all rooms in memory
var Mansion = []*Room{}

type Room struct {
	Id           string
	Playlist     Playlist
	PlayedTracks []Track
	CurrentTrack chan Track
	Players      map[string]*Player
}

// Create a new Room with a random uuid as id
// @param playlist: The playlist to play in the room
func NewRoom(playlist Playlist) *Room {
	return NewRoomWithId(uuid.New().String(), playlist)
}

// Create a new Room with the specified id
func NewRoomWithId(id string, playlist Playlist) *Room {
	newRoom := &Room{
		Id:           id,
		Playlist:     playlist,
		PlayedTracks: make([]Track, 0, len(playlist.Tracks.Items)),
		CurrentTrack: make(chan Track),
		Players:      make(map[string]*Player),
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

		r.PlayedTracks = append(r.PlayedTracks, newTrack)

		// Create a new set of results for each player in the room
		for _, player := range r.Players {
			newGuess := GuessResult{
				Title:   false,
				Artists: make(map[string]bool, len(newTrack.Artists)),
			}
			for _, artist := range newTrack.Artists {
				newGuess.Artists[artist.Name] = false
			}
			player.Guesses[newTrack.Name] = &newGuess
		}

		r.CurrentTrack <- newTrack
		time.Sleep(TRACK_DURATION)
	}
}

func (r *Room) GuessResult(playerId, guess string) *GuessResult {
	currentTrack := r.PlayedTracks[len(r.PlayedTracks)-1]

	normalizedGuess := utils.Normalize(guess)
	normalizedTitle := utils.Normalize(currentTrack.Name)

	guessLen := utf8.RuneCountInString(normalizedGuess)
	titleLen := utf8.RuneCountInString(normalizedTitle)
	score := fuzzy.LevenshteinDistance(normalizedGuess, normalizedTitle)

	normalizedScore := 100 * (float32(max(guessLen, titleLen)) - float32(score)) / float32(max(guessLen, titleLen))

	player := r.Players[playerId]

	oldGuessResult := player.Guesses[currentTrack.Name]

	newGuessResult := GuessResult{
		Title:   normalizedScore >= GUESS_VALIDITY_THRESHOLD || oldGuessResult.Title,
		Artists: make(map[string]bool, len(currentTrack.Artists)),
	}
	for _, artist := range currentTrack.Artists {
		newGuessResult.Artists[artist.Name] = oldGuessResult.Artists[artist.Name]
	}

	player.Guesses[currentTrack.Name] = &newGuessResult
	return &newGuessResult
}

func (r *Room) AddPlayer(player Player) {
	log.Infof("Player %s joined room %s", player.Id, r.Id)

	r.Players[player.Id] = &player
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
