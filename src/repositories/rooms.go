package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v3/log"
	_ "github.com/mattn/go-sqlite3"

	"lcor.io/songs/src/models"
	"lcor.io/songs/src/services"
)

const dev_file string = "rooms.db"

type RoomRepository struct {
	db *sql.DB
}

func (repo *RoomRepository) CreateRoom(opts services.RoomOpts, playlist models.Playlist) (string, error) {
	rollback := func() {
		repo.db.Exec("ROLLBACK;")
	}

	if _, err := repo.db.Exec("BEGIN TRANSACTION;"); err != nil {
		return "", fmt.Errorf("Error starting transaction: %v", err)
	}

	// Inserting room
	roomRes, err := repo.db.Exec("INSERT INTO rooms DEFAULT VALUES;")
	if err != nil {
		rollback()
		return "", fmt.Errorf("Error inserting room: %v", err)
	}

	newRoomId, err := roomRes.LastInsertId()
	if err != nil {
		rollback()
		return "", fmt.Errorf("Error getting playlist id: %v", err)
	}

	log.Infof("New room id: %d", newRoomId)

	// Inserting room options
	if _, err := repo.db.Exec(`
    INSERT INTO room_opts
      (track_duration, guess_vality_threshold, guess_partial_threshold, max_players, room_id)
    VALUES 
      (?, ? ,? ,?, ?);`,
		opts.TrackDuration.Seconds(),
		opts.GuessValidityThreshold,
		opts.GuessPartialThreshold,
		opts.MaxPlayerNumber,
		newRoomId,
	); err != nil {
		rollback()
		return "", fmt.Errorf("Error inserting room options: %v", err)
	}

	// Inserting playlist
	playlistRes, err := repo.db.Exec(`
    INSERT INTO playlists
      (name, client_id, room_id)
    VALUES
      (?, ?, ?);
    `,
		playlist.Name,
		playlist.ID,
		newRoomId)
	if err != nil {
		rollback()
		return "", fmt.Errorf("Error inserting playlist: %v", err)
	}

	newPlaylistId, err := playlistRes.LastInsertId()
	if err != nil {
		rollback()
		return "", fmt.Errorf("Error getting playlist id: %v", err)
	}

	// Inserting tracks
	for _, track := range playlist.Tracks {

		// Check if artist exists
		var artistId int64
		err := repo.db.QueryRow("SELECT id FROM artists WHERE name = ?", track.Artists[0].Name).Scan(&artistId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			rollback()
			return "", fmt.Errorf("Error checking artist: %v", err)
		}

		if errors.Is(err, sql.ErrNoRows) {
			newArtistQuery, err := repo.db.Exec(`INSERT INTO artists (name, client, link) VALUES (?,?,?);`, track.Artists[0].Name, track.Client, track.Artists[0].Link)
			if err != nil {
				rollback()
				return "", fmt.Errorf("Error inserting artist: %v", err)
			}
			artistId, err = newArtistQuery.LastInsertId()
			if err != nil {
				rollback()
				return "", fmt.Errorf("Error getting artist id: %v", err)
			}
		}

		newTrackQuery, err := repo.db.Exec(`
      INSERT INTO tracks
        (name, client, preview, link, playlist_id)
      VALUES
        (?, ?, ?, ?, ?);
      `,
			track.Name,
			track.Client,
			track.PreviewUrl,
			"",
			newPlaylistId,
		)
		if err != nil {
			rollback()
			return "", fmt.Errorf("Error inserting track: %v", err)
		}
		trackId, err := newTrackQuery.LastInsertId()
		if err != nil {
			rollback()
			return "", fmt.Errorf("Error getting track id: %v", err)
		}

		_, err = repo.db.Exec(`
      INSERT INTO artists_to_tracks
        (artist_id, track_id)
      VALUES
        (?, ?);
      `, artistId, trackId)
		if err != nil {
			rollback()
			return "", fmt.Errorf("Error inserting artist to track: %v", err)
		}
	}

	if _, err = repo.db.Exec("COMMIT;"); err != nil {
		rollback()
		return "", fmt.Errorf("Error commiting transaction: %v", err)
	}

	return "", nil
}

func (repo *RoomRepository) GetRoom(id int) (services.Room, error) {
	panic("not implemented")
}

func (repo *RoomRepository) GetRooms() ([]services.Room, error) {
	panic("not implemented")
}

func GetLocalRepository() *RoomRepository {
	db, err := sql.Open("sqlite3", dev_file)
	if err != nil {
		panic(err)
	}

	init, err := os.ReadFile("../src/utils/init.sql")
	if err != nil {
		panic(fmt.Errorf("Error reading init.sql: %v", err))
	}

	if _, err := db.Exec(string(init)); err != nil {
		panic(fmt.Errorf("Error executing init.sql: %v", err))
	}

	return &RoomRepository{db}
}

func InitDb() {
	db, err := sql.Open("sqlite3", dev_file)
	if err != nil {
		panic(err)
	}

	init, err := os.ReadFile("init.sql")
	if err != nil {
		panic(fmt.Errorf("Error reading init.sql: %v", err))
	}

	if _, err := db.Exec(string(init)); err != nil {
		panic(fmt.Errorf("Error executing init.sql: %v", err))
	}
}
