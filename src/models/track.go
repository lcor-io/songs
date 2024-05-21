package models

type Client string

const (
	Spotify Client = "spotify"
)

type Track struct {
  ID         string
	Artists    []Artist
	Name       string
	Client     Client
	Link       string
	PreviewUrl string
	Image      Image
}
