PRAGMA foreign_keys = ON;

DROP TABLE IF EXISTS rooms;

CREATE TABLE rooms (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  likes INTEGER DEFAULT 0,
  created_on DATE DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS room_opts;

CREATE TABLE room_opts (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  track_duration REAL NOT NULL CHECK (track_duration > 0),
  guess_vality_threshold INTEGER NOT NULL,
  guess_partial_threshold INTEGER NOT NULL,
  max_players INTEGER NOT NULL CHECK (max_players > 0),
  room_id INTEGER NOT NULL,
  CONSTRAINT fk_room FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS playlists;

CREATE TABLE playlists (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  client_id TEXT NOT NULL,
  room_id INTEGER NOT NULL,
  CONSTRAINT fk_room FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS artists_to_tracks;

CREATE TABLE artists_to_tracks (
  artist_id INTEGER,
  track_id INTEGER,
  PRIMARY KEY (artist_id, track_id),
  CONSTRAINT fk_artist FOREIGN KEY (artist_id) REFERENCES artists (id) ON DELETE CASCADE,
  CONSTRAINT fk_track FOREIGN KEY (track_id) REFERENCES tracks (id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS artists;

CREATE TABLE artists (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  client TEXT NOT NULL,
  link TEXT
);

DROP TABLE IF EXISTS tracks;

CREATE TABLE tracks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  client TEXT NOT NULL,
  preview TEXT NOT NULL,
  link TEXT,
  playlist_id INTEGER NOT NULL,
  CONSTRAINT fk_playlist FOREIGN KEY (playlist_id) REFERENCES playlists (id) ON DELETE CASCADE
);
