package services

import (
	"maps"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/google/uuid"

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
	score   float32
}

type RoomPlayer struct {
	Id       string
	PlayerId string
	Guesses  map[string]*GuessResult
	score    float32
}

type Room struct {
	Id           string
	Playlist     *Playlist
	PlayedTracks []Track
	CurrentTrack chan Track
	Players      map[string]*RoomPlayer
	Scores       chan []struct {
		Id    string
		Score float32
	}
	done   chan bool
	ticker *time.Ticker
	mu     sync.Mutex
}

func NewRoom(playlist *Playlist) *Room {
	return &Room{
		Id:           uuid.NewString(),
		Playlist:     playlist,
		PlayedTracks: make([]Track, 0, len(playlist.Tracks.Items)),
		CurrentTrack: make(chan Track, len(playlist.Tracks.Items)),
		Players:      make(map[string]*RoomPlayer),
		Scores: make(chan []struct {
			Id    string
			Score float32
		}),
		done: make(chan bool),
	}
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
		r.PlayedTracks = append(r.PlayedTracks, newTrack)
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
		score:   oldGuessResult.score,
	}
	var newGuessScore float32 = 0
	for _, guess := range guessCombinations {
		if newGuessResult.Title != Valid {
			score := utils.GetScore(guess, normalizedTitle)
			switch {
			case score >= GUESS_VALIDITY_THRESHOLD:
				newGuessResult.Title = Valid
				newGuessScore += 1
			case score >= GUESS_PARTIAL_THRESHOLD:
				newGuessResult.Title = Partial
				newGuessScore += score
			default:
				newGuessResult.Title = Invalid
			}
		} else {
			newGuessScore += 1
		}
		for _, artist := range normalizedArtists {
			if newGuessResult.Artists[artist] != Valid {
				score := utils.GetScore(guess, artist)
				switch {
				case score >= GUESS_VALIDITY_THRESHOLD:
					newGuessResult.Artists[artist] = Valid
					newGuessScore += 1
				case score >= GUESS_PARTIAL_THRESHOLD:
					newGuessResult.Artists[artist] = Partial
					newGuessScore += score
				default:
					newGuessResult.Artists[artist] = Invalid
				}
			} else {
				newGuessScore += 1
			}
		}

		if newGuessScore > oldGuessResult.score {
			newGuessResult.score = newGuessScore
			player.score += (newGuessScore - oldGuessResult.score)
		}
	}

	// Update the score for all players in the room
	if newGuessResult.score > oldGuessResult.score {
		scores := make([]struct {
			Id    string
			Score float32
		}, len(r.Players))
		for _, player := range r.Players {
			scores = append(scores, struct {
				Id    string
				Score float32
			}{player.Id, player.score})
		}
		slices.SortStableFunc(scores, func(a, b struct {
			Id    string
			Score float32
		},
		) int {
			return int(a.Score - b.Score)
		})
		r.Scores <- scores
	}

	player.Guesses[currentTrack.Name] = &newGuessResult
	return &newGuessResult
}

func (r *Room) AddPlayer(player RoomPlayer) {
	log.Infof("Player %s joined room %s", player.Id, r.Id)
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Players[player.Id] = &player
	if len(r.Players) == 1 {
		go r.Launch()
	}
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
