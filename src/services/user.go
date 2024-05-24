package services

import (
	"errors"

	"github.com/google/uuid"
	"lcor.io/songs/src/models"
)

var users = map[string]*models.User{}

func UserExists(id string) bool {
	_, exists := users[id]
	return exists
}

func GetUser(id string) (*models.User, error) {
	if user, exists := users[id]; exists {
		return user, nil
	}
	return nil, errors.New("User not found")
}

func GetUsers() map[string]*models.User {
	return users
}

func CreateUser(name string) *models.User {
	user := models.User{
		ID:   uuid.NewString(),
		Name: name,
	}
	users[user.ID] = &user
	return &user
}
