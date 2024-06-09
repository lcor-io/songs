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

	"lcor.io/songs/src/models"
	"lcor.io/songs/src/utils"
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

type Player struct {
	Id       string
	Name     string
	PlayerId string
	Guesses  map[string]*GuessResult
	score    float32
	Nonce    uint8 // Used to reconnect a user if a leave a room
}

type RoomOpts struct {
	TrackDuration          time.Duration
	GuessValidityThreshold int8
	GuessPartialThreshold  int8
	MaxPlayerNumber        int8
}

type roomOptFunc func(*RoomOpts)

func defaultOpts() RoomOpts {
	return RoomOpts{
		TrackDuration:          30 * time.Second,
		GuessValidityThreshold: 80,
		GuessPartialThreshold:  50,
		MaxPlayerNumber:        15,
	}
}

func WithTrackDuration(d time.Duration) roomOptFunc {
	return func(o *RoomOpts) {
		o.TrackDuration = d
	}
}

func WithGuessValidityThreshold(t int8) roomOptFunc {
	return func(o *RoomOpts) {
		o.GuessValidityThreshold = t
	}
}

func WithGuessPartialThreshold(t int8) roomOptFunc {
	return func(o *RoomOpts) {
		o.GuessPartialThreshold = t
	}
}

func WithMaxPlayerNumber(n int8) roomOptFunc {
	return func(o *RoomOpts) {
		o.MaxPlayerNumber = n
	}
}

type Room struct {
	Id string

	opts RoomOpts

	Playlist     *models.Playlist
	PlayedTracks []models.Track
	CurrentTrack chan models.Track
	Players      map[string]*Player
	Scores       chan []struct {
		Id    string
		Score float32
	}

	connectionNumber uint8
	done             chan bool
	ticker           *time.Ticker
	mu               sync.Mutex
}

func NewRoom(playlist models.Playlist, opts ...roomOptFunc) *Room {
	opt := defaultOpts()

	// Apply room options in order
	for _, fn := range opts {
		fn(&opt)
	}

	return &Room{
		Id: uuid.NewString(),

		opts: opt,

		Playlist:     &playlist,
		PlayedTracks: make([]models.Track, 0, len(playlist.Tracks)),
		CurrentTrack: make(chan models.Track, opt.MaxPlayerNumber),
		Players:      make(map[string]*Player),
		Scores: make(chan []struct {
			Id    string
			Score float32
		}, opt.MaxPlayerNumber),

		connectionNumber: 0,
		done:             make(chan bool),
	}
}

func (r *Room) Launch() {
	defer func() {
		for _, player := range r.Players {
			r.RemovePlayer(player.Id, player.Nonce)
		}
	}()

	playlistTracks := r.Playlist.Tracks
	r.ticker = time.NewTicker(r.opts.TrackDuration)

	processNewTrack := func() {
		// Select a track random track from playlist not in already played tracks
		newTrack := playlistTracks[rand.Intn(len(playlistTracks))]
		newTrackAlreadyPlayed := slices.ContainsFunc(r.PlayedTracks, func(t models.Track) bool {
			return t.Name == newTrack.Name
		})
		for newTrackAlreadyPlayed {
			newTrack = playlistTracks[rand.Intn(len(playlistTracks))]
			newTrackAlreadyPlayed = slices.ContainsFunc(r.PlayedTracks, func(t models.Track) bool {
				return t.Name == newTrack.Name
			})
		}

		r.PlayedTracks = append(r.PlayedTracks, newTrack)

		// Create a new set of results for each player in the room and send the
		// new track
		r.mu.Lock()
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
		for i := uint8(0); i < r.connectionNumber; i++ {
			r.CurrentTrack <- newTrack
		}
		r.mu.Unlock()
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
	for _, guess := range guessCombinations {
		var newGuessScore float32 = 0
		if newGuessResult.Title != Valid {
			score := utils.GetScore(guess, normalizedTitle)
			switch {
			case score >= float32(r.opts.GuessValidityThreshold):
				// Add a bonus for the first player to find the title
				alreadyFound := 0
				r.mu.Lock()
				for _, player := range r.Players {
					if player.Id == player.Id {
						continue
					}
					currentGuess := player.Guesses[currentTrack.Name]
					if currentGuess.Title == Valid {
						alreadyFound++
					}
				}
				r.mu.Unlock()
				switch alreadyFound {
				case 0:
					newGuessScore += 50
				case 1:
					newGuessScore += 25
				case 2:
					newGuessScore += 15
				}
				newGuessResult.Title = Valid
				newGuessScore += 100
			case score >= float32(r.opts.GuessPartialThreshold):
				newGuessResult.Title = Partial
				newGuessScore += score
			default:
				newGuessResult.Title = Invalid
			}

		} else {
			newGuessScore += 100
		}
		for _, artist := range normalizedArtists {
			if newGuessResult.Artists[artist] != Valid {
				score := utils.GetScore(guess, artist)
				switch {
				case score >= float32(r.opts.GuessValidityThreshold):
					// Add a bonus for the first player to find the title
					alreadyFound := 0
					r.mu.Lock()
					for _, player := range r.Players {
						if player.Id == player.Id {
							continue
						}
						currentGuess := player.Guesses[currentTrack.Name]
						if currentGuess.Title == Valid {
							alreadyFound++
						}
					}
					r.mu.Unlock()
					switch alreadyFound {
					case 0:
						newGuessScore += 50
					case 1:
						newGuessScore += 25
					case 2:
						newGuessScore += 15
					}
					newGuessResult.Artists[artist] = Valid
					newGuessScore += 100
				case score >= float32(r.opts.GuessPartialThreshold):
					newGuessResult.Artists[artist] = Partial
					newGuessScore += score
				default:
					newGuessResult.Artists[artist] = Invalid
				}
			} else {
				newGuessScore += 100
			}
		}

		if newGuessScore > newGuessResult.score {
			newGuessResult.score = newGuessScore
		}
	}

	player.score += (newGuessResult.score - oldGuessResult.score)

	// Update the score for all players in the room
	if newGuessResult.score > oldGuessResult.score {
		scores := make([]struct {
			Id    string
			Score float32
		}, 0, len(r.Players))
		for _, player := range r.Players {
			scores = append(scores, struct {
				Id    string
				Score float32
			}{player.Name, player.score})
		}
		slices.SortFunc(scores, func(a, b struct {
			Id    string
			Score float32
		},
		) int {
			return int(b.Score - a.Score)
		})

		// Send the score to all the players
		for i := uint8(0); i < r.connectionNumber; i++ {
			r.Scores <- scores
		}
	}

	player.Guesses[currentTrack.Name] = &newGuessResult
	return &newGuessResult
}

func (r *Room) AddPlayer(user *models.User) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.connectionNumber += 1

	// The player is already in the room, we just need to update the missing guesses
	if player, exists := r.Players[user.ID]; exists {
		log.Infof("Player %s reconnected to room %s", user.ID, r.Id)
		for _, track := range r.PlayedTracks {
			if _, exists := player.Guesses[track.Name]; !exists {
				player.Guesses[track.Name] = &GuessResult{
					Title:   Invalid,
					Artists: make(map[string]ResultValidity, len(track.Artists)),
				}
			}
		}
		return
	}

	log.Infof("Player %s joined room %s", user.ID, r.Id)

	player := Player{
		Id:       uuid.NewString(),
		Name:     user.Name,
		PlayerId: user.ID,
	}

	// Add every guesses for each elapsed tracks
	guesses := make(map[string]*GuessResult)
	for _, track := range r.PlayedTracks {
		guesses[track.Name] = &GuessResult{
			Title:   Invalid,
			Artists: make(map[string]ResultValidity, len(track.Artists)),
		}
	}
	player.Guesses = guesses

	r.Players[user.ID] = &player

	// If the room was empty before, we can start it
	if len(r.Players) == 1 {
		go r.Launch()
	}
}

func (r *Room) RemovePlayer(id string, nonce uint8) {
	r.connectionNumber -= 1

	if playerNonce := r.Players[id].Nonce; playerNonce != nonce {
		return
	}

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
