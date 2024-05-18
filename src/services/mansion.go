package services

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

type mansion struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

var Mansion = mansion{rooms: map[string]*Room{}}

func (m *mansion) GetAll() map[string]*Room {
	return m.rooms
}

func (m *mansion) NewRoom(playlist Playlist) *Room {
	return m.NewRoomWithId(uuid.New().String(), playlist)
}

func (m *mansion) NewRoomWithId(id string, playlist Playlist) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	newRoom := Room{
		Id:           id,
		Playlist:     playlist,
		PlayedTracks: make([]Track, 0, len(playlist.Tracks.Items)),
		CurrentTrack: make(chan Track, len(playlist.Tracks.Items)),
		Players:      make(map[string]*Player),
		done:         make(chan bool),
	}

	m.rooms[id] = &newRoom
	return &newRoom
}

func (m *mansion) GetRoom(id string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, exists := m.rooms[id]; exists {
		return room, nil
	}
	return nil, errors.New("Room not found")
}

func (m *mansion) RemoveRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.rooms, id)
}
