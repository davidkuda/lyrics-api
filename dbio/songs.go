package dbio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/davidkuda/lyricsapi/models"
)

var ErrSongDoesNotExist = errors.New("Song does not exist")

func ListSongs(db sql.DB) models.Songs {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	query := "SELECT artist, name, id FROM songs ORDER BY artist ASC;"
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	var songs models.Songs
	for rows.Next() {
		song := models.Song{}
		rows.Scan(&song.Artist, &song.Name, &song.ID)
		songs = append(songs, song)
	}

	return songs
}

func GetSong(songID string, db sql.DB, l log.Logger) (models.Song, error) {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	song := models.Song{}

	query := `
		SELECT
			id,
			artist,
			name,
			text,
			chords,
			copyright
		FROM songs
		WHERE id = $1`

	row := conn.QueryRowContext(context.Background(), query, songID)

	if row.Err(); err != nil {
		l.Println("conn.QueryRow", err)
		return song, errors.New("QueryNotSuccesful")
	}

	row.Scan(
		&song.ID,
		&song.Artist,
		&song.Name,
		&song.Text,
		&song.Chords,
		&song.Copyright,
	)

	if len(song.Name) == 0 {
		return song, ErrSongDoesNotExist
	}

	return song, nil
}

func CreateSong(s *models.Song, db sql.DB, l log.Logger) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := `
		INSERT INTO songs (
			id,
			artist,
			name,
			text,
			chords,
			copyright
		) VALUES ($1, $2, $3, $4, $5, $6);`

	if _, err := conn.ExecContext(
		ctx, query, s.ID, s.Artist, s.Name, s.Text, s.Chords, s.Copyright,
	); err != nil {
		l.Println("conn.ExecContext:", err)
		return err
	}

	return nil
}

func DeleteSong(songID string, db sql.DB, l log.Logger) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := "DELETE FROM songs WHERE song_id = $1;"

	res, err := conn.ExecContext(ctx, query, songID)
	if err != nil {
		l.Println("conn.ExecContext:", err)
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		l.Println("res.RowsAffected:", err)
		return err
	}
	if n == 0 {
		return errors.New("Delete failed: Did not find song with id " + songID)
	}

	return nil
}
