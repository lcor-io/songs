package services

import (
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type Track struct {
	Artists []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		Id     string `json:"id"`
		Images []struct {
			Height int    `json:"height"`
			Url    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"images"`
		Name string `json:"name"`
		Type string `json:"type"`
		Uri  string `json:"uri"`
	} `json:"artists"`
	Name       string `json:"name"`
	PreviewUrl string `json:"preview_url"`
}
type Playlist struct {
	Collaborative bool   `json:"collaborative"`
	Description   string `json:"description"`
	ExternalUrls  struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href   string `json:"href"`
	Id     string `json:"id"`
	Images []struct {
		Height int    `json:"height"`
		Url    string `json:"url"`
		Width  int    `json:"width"`
	} `json:"images"`
	Name  string `json:"name"`
	Owner struct {
		DisplayName string `json:"display_name"`
		Followers   struct {
			Href  string `json:"href"`
			Total int    `json:"total"`
		} `json:"followers"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		Id   string `json:"id"`
		Type string `json:"type"`
		Uri  string `json:"uri"`
	} `json:"owner"`
	Public     bool   `json:"public"`
	SnapshotId string `json:"snapshot_id"`
	Tracks     struct {
		Href  string `json:"href"`
		Total int    `json:"total"`
		Items []struct {
			Track Track `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
	Type string `json:"type"`
	Uri  string `json:"uri"`
}
type SpotifyPlaylistResult struct {
	Message   string `json:"message"`
	Playlists struct {
		Href     string     `json:"href"`
		Limit    int        `json:"limit"`
		Next     string     `json:"next"`
		Offset   int        `json:"offset"`
		Previous string     `json:"previous"`
		Total    int        `json:"total"`
		Items    []Playlist `json:"items"`
	} `json:"playlists"`
}
type credentialsResult struct {
	AccessToken string `json:"access_token"`
	Token_type  string `json:"token_type"`
	Expires_in  int    `json:"expires_in"`
}

type SpotifyService struct {
	credentials string
	access      *credentialsResult
	client      *resty.Client
}

const (
	BASE_URL      = "https://api.spotify.com/v1"
	BASE_URL_AUTH = "https://accounts.spotify.com/api/token"
)

func (s SpotifyService) GetFeaturedPlaylist() SpotifyPlaylistResult {
	res := &SpotifyPlaylistResult{}
	s.client.R().SetResult(res).Get(BASE_URL + "/browse/featured-playlists?limit=10")
	return *res
}

func (s SpotifyService) GetPlaylist(id string) Playlist {
	res := &Playlist{}
	s.client.R().SetResult(res).Get(BASE_URL + "/playlists/" + id)
	return *res
}

func (s SpotifyService) refreshToken() {
	log.Println("Refreshing Spotify token...")
	_, err := s.client.R().
		SetHeader("Authorization", "Basic "+s.credentials).
		SetFormData(map[string]string{
			"grant_type": "client_credentials",
		}).
		SetResult(&s.access).
		Post(BASE_URL_AUTH)
	if err != nil {
		log.Panicln("Could not refresh Spotify token", err)
	}
	log.Println("Spotify token refreshed")
}

func Spotify(credentials string) *SpotifyService {
	spotify := SpotifyService{}
	client := resty.New().
		SetRetryCount(1).
		OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
			if spotify.access.AccessToken != "" {
				r.SetAuthToken(spotify.access.AccessToken)
			}
			return nil
		}).
		OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			if r.StatusCode() == http.StatusUnauthorized {
				spotify.refreshToken()
			}
			return nil
		}).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusUnauthorized
		})
	spotify.credentials = credentials
	spotify.client = client
	spotify.access = &credentialsResult{
		AccessToken: "",
		Token_type:  "",
		Expires_in:  0,
	}
	return &spotify
}
