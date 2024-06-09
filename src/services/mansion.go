package services

import (
	"errors"
	"sync"

	"lcor.io/songs/src/models"
)

type mansion struct {
	mu          sync.Mutex
	activeRooms map[string]*Room
}

var Mansion = mansion{activeRooms: map[string]*Room{}}

func (m *mansion) GetAll() map[string]*Room {
	return m.activeRooms
}

func (m *mansion) NewRoom(playlist models.Playlist) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	newRoom := NewRoom(playlist)
	m.activeRooms[newRoom.Id] = newRoom
	return newRoom
}

func (m *mansion) GetRoom(id string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, exists := m.activeRooms[id]; exists {
		return room, nil
	}
	return nil, errors.New("Room not found")
}

func (m *mansion) RemoveRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.activeRooms, id)
}
