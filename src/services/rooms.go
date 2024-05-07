package services

import (
	"errors"
	"maps"
	"math/rand"
	"slices"
	"strings"
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
				newGuess.Artists[utils.Normalize(artist.Name)] = false
			}
			player.Guesses[newTrack.Name] = &newGuess
		}

		r.CurrentTrack <- newTrack
		time.Sleep(TRACK_DURATION)
	}
}

func (r *Room) GuessResult(playerId, guess string) *GuessResult {
	currentTrack := r.PlayedTracks[len(r.PlayedTracks)-1]

	// Normalize inputs for comparison
	normalizedGuess := utils.Normalize(guess)
	normalizedTitle := utils.Normalize(currentTrack.Name)
	normalizedArtists := make([]string, len(currentTrack.Artists))
	for i, artist := range currentTrack.Artists {
		normalizedArtists[i] = utils.Normalize(artist.Name)
	}

	// The player can guess the artists and the title at the same, so we need
	// to check every possible combination of the input
	guessCombinations := utils.Permutations(strings.Fields(normalizedGuess))

	player := r.Players[playerId]
	oldGuessResult := player.Guesses[currentTrack.Name]

	// Match inputs against title and artists
	newGuessResult := GuessResult{
		Title:   oldGuessResult.Title,
		Artists: maps.Clone(oldGuessResult.Artists),
	}
	for _, guess := range guessCombinations {
		if !newGuessResult.Title {
			guessLen := utf8.RuneCountInString(guess)
			titleLen := utf8.RuneCountInString(normalizedTitle)
			score := fuzzy.LevenshteinDistance(guess, normalizedTitle)
			normalizedScore := 100 * (float32(max(guessLen, titleLen)) - float32(score)) / float32(max(guessLen, titleLen))
			newGuessResult.Title = normalizedScore >= GUESS_VALIDITY_THRESHOLD
		}
		for _, artist := range normalizedArtists {
			if !newGuessResult.Artists[artist] {
				guessLen := utf8.RuneCountInString(guess)
				artistLen := utf8.RuneCountInString(artist)
				score := fuzzy.LevenshteinDistance(guess, artist)
				normalizedScore := 100 * (float32(max(guessLen, artistLen)) - float32(score)) / float32(max(guessLen, artistLen))
				newGuessResult.Artists[artist] = normalizedScore >= GUESS_VALIDITY_THRESHOLD
			}
		}
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
