package services

import (
	"maps"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3/log"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"lcor.io/songs/src/utils"
)

const (
	TRACK_DURATION           = 30 * time.Second
	GUESS_VALIDITY_THRESHOLD = 85
	GUESS_PARTIAL_THRESHOLD  = 70
)

type ResultValidity uint8

const (
	Invalid ResultValidity = iota
	Partial
	Valid
)

type GuessResult struct {
	Title   ResultValidity
	Artists map[string]ResultValidity
}

type Player struct {
	Id      string
	Guesses map[string]*GuessResult
}

type Room struct {
	Id           string
	Playlist     Playlist
	PlayedTracks []Track
	CurrentTrack chan Track
	Players      map[string]*Player
	done         chan bool
	ticker       *time.Ticker
	mu           sync.Mutex
}

func (r *Room) Launch() {
	defer func() {
		for _, player := range r.Players {
			r.RemovePlayer(player.Id)
		}
	}()

	playlistTracks := r.Playlist.Tracks.Items
	r.ticker = time.NewTicker(TRACK_DURATION)

	processNewTrack := func() {
		// Select a track random track from playlist not in already played tracks
		newTrack := playlistTracks[rand.Intn(len(playlistTracks))].Track
		newTrackAlreadyPlayed := slices.ContainsFunc(r.PlayedTracks, func(t Track) bool {
			return t.Name == newTrack.Name
		})
		for newTrackAlreadyPlayed {
			newTrack = playlistTracks[rand.Intn(len(playlistTracks))].Track
			newTrackAlreadyPlayed = slices.ContainsFunc(r.PlayedTracks, func(t Track) bool {
				return t.Name == newTrack.Name
			})
		}

		r.PlayedTracks = append(r.PlayedTracks, newTrack)

		// Create a new set of results for each player in the room
		for _, player := range r.Players {
			newGuess := GuessResult{
				Title:   Invalid,
				Artists: make(map[string]ResultValidity, len(newTrack.Artists)),
			}
			for _, artist := range newTrack.Artists {
				newGuess.Artists[utils.Normalize(artist.Name)] = Invalid
			}
			player.Guesses[newTrack.Name] = &newGuess
		}

		r.CurrentTrack <- newTrack
	}

	for i := 0; i < len(playlistTracks); i++ {
		if i == 0 {
			processNewTrack()
			continue
		}
		select {
		case <-r.done:
			return
		case <-r.ticker.C:
			processNewTrack()
		}
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
		if newGuessResult.Title != Valid {
			guessLen := utf8.RuneCountInString(guess)
			titleLen := utf8.RuneCountInString(normalizedTitle)
			score := fuzzy.LevenshteinDistance(guess, normalizedTitle)
			normalizedScore := 100 * (float32(max(guessLen, titleLen)) - float32(score)) / float32(max(guessLen, titleLen))
			if normalizedScore >= GUESS_VALIDITY_THRESHOLD {
				newGuessResult.Title = Valid
			} else if normalizedScore >= GUESS_PARTIAL_THRESHOLD {
				newGuessResult.Title = Partial
			} else {
				newGuessResult.Title = Invalid
			}
		}
		for _, artist := range normalizedArtists {
			if newGuessResult.Artists[artist] != Valid {
				guessLen := utf8.RuneCountInString(guess)
				artistLen := utf8.RuneCountInString(artist)
				score := fuzzy.LevenshteinDistance(guess, artist)
				normalizedScore := 100 * (float32(max(guessLen, artistLen)) - float32(score)) / float32(max(guessLen, artistLen))
				if normalizedScore >= GUESS_VALIDITY_THRESHOLD {
					newGuessResult.Artists[artist] = Valid
				} else if normalizedScore >= GUESS_PARTIAL_THRESHOLD {
					newGuessResult.Artists[artist] = Partial
				} else {
					newGuessResult.Artists[artist] = Invalid
				}
			}
		}
	}

	player.Guesses[currentTrack.Name] = &newGuessResult
	return &newGuessResult
}

func (r *Room) AddPlayer(player Player) {
	log.Infof("Player %s joined room %s", player.Id, r.Id)
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Players[player.Id] = &player
}

func (r *Room) RemovePlayer(id string) {
	log.Infof("Player %s leaved room %s", id, r.Id)

	r.mu.Lock()
	delete(r.Players, id)
	r.mu.Unlock()

	// The room is empty, remove it
	if len(r.Players) == 0 {
		log.Infof("Room %s is empty, removing it", r.Id)
		r.ticker.Stop()
		r.done <- true
		Mansion.RemoveRoom(r.Id)
		r = nil
	}
}
