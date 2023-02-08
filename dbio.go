package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrSongDoesNotExist = errors.New("Song does not exist")

// returns a pool of connections to the postgres db according to the args
func getDatabaseConn(dbAddr, dbName, dbUser, dbPassword string) (*sql.DB, error) {
	// "data source name": string of the url to the database
	dsn := url.URL{
		Scheme: "postgres",
		Host:   dbAddr,
		User:   url.UserPassword(dbUser, dbPassword),
		Path:   dbName,
	}
	return sql.Open("pgx", dsn.String())
}

func ListSongs(cfg appConfig) Songs {
	ctx := context.Background()
	conn, err := cfg.db.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	query := "SELECT artist, song_name FROM songs ORDER BY artist ASC;"
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	var songs Songs
	var artist, song string
	for rows.Next() {
		rows.Scan(&artist, &song)
		song := Song{Artist: artist, SongName: song}
		songs = append(songs, song)
	}

	return songs
}

func GetSong(songName string, cfg appConfig) (Song, error) {
	ctx := context.Background()
	conn, err := cfg.db.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	song := Song{}

	query := `
		SELECT
			artist,
			song_name,
			song_text,
			chords,
			copyright
		FROM songs
		WHERE song_name = $1`

	row := conn.QueryRowContext(context.Background(), query, songName)

	if row.Err(); err != nil {
		cfg.logger.Println("conn.QueryRow", err)
		return song, errors.New("QueryNotSuccesful")
	}

	row.Scan(
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
