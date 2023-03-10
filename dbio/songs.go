package dbio

import (
	"context"
	"errors"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/davidkuda/lyricsapi/config"
	"github.com/davidkuda/lyricsapi/models"
)

var ErrSongDoesNotExist = errors.New("Song does not exist")

func ListSongs(cfg config.AppConfig) models.Songs {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	query := "SELECT artist, song_name, song_id FROM songs ORDER BY artist ASC;"
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	var songs models.Songs
	var artist, song, song_id string
	for rows.Next() {
		rows.Scan(&artist, &song, &song_id)
		song := models.Song{Artist: artist, SongName: song, SongID: song_id}
		songs = append(songs, song)
	}

	return songs
}

func GetSong(songID string, cfg config.AppConfig) (models.Song, error) {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	song := models.Song{}

	query := `
		SELECT
			song_id,
			artist,
			song_name,
			song_text,
			chords,
			copyright
		FROM songs
		WHERE song_id = $1`

	row := conn.QueryRowContext(context.Background(), query, songID)

	if row.Err(); err != nil {
		cfg.Logger.Println("conn.QueryRow", err)
		return song, errors.New("QueryNotSuccesful")
	}

	row.Scan(
		&song.SongID,
		&song.Artist,
		&song.SongName,
		&song.SongText,
		&song.Chords,
		&song.Copyright,
	)

	if len(song.SongName) == 0 {
		return song, ErrSongDoesNotExist
	}

	return song, nil
}

func CreateSong(s *models.Song, cfg config.AppConfig) error {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := `
		INSERT INTO songs (
			song_id,
			artist,
			song_name,
			song_text,
			chords,
			copyright
		) VALUES ($1, $2, $3, $4, $5, $6);`

	if _, err := conn.ExecContext(
		ctx, query, s.SongID, s.Artist, s.SongName, s.SongText, s.Chords, s.Copyright,
	); err != nil {
		cfg.Logger.Println("conn.ExecContext:", err)
		return err
	}

	return nil
}

func DeleteSong(songID string, cfg config.AppConfig) error {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := "DELETE FROM songs WHERE song_id = $1;"

	res, err := conn.ExecContext(ctx, query, songID)
	if err != nil {
		cfg.Logger.Println("conn.ExecContext:", err)
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		cfg.Logger.Println("res.RowsAffected:", err)
		return err
	}
	if n == 0 {
		return errors.New("Delete failed: Did not find song with id " + songID)
	}

	return nil
}
