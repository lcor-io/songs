package models

type Playlist struct {
	ID     string
	Name   string
	Tracks []Track
	Link   string
	Image  Image
}
