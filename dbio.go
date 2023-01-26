package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func ListSongs() Songs {
	os.Setenv("DATABASE_URL", "postgres://lyricsapi:lyricsapi@localhost:5432/lyricsapi")
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	query := "SELECT artist, song_name FROM songs ORDER BY artist ASC;"
	// ? How to know how to use conn.Query?
	rows, err := conn.Query(context.Background(), query)
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

func GetSong(songName string) Song {
	// TODO: Validate input, avoid SQLInjection, check against all available songs, store all songs in memory for fast check
	os.Setenv("DATABASE_URL", "postgres://lyricsapi:lyricsapi@localhost:5432/lyricsapi")
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	song := Song{}

	query := fmt.Sprintf(
		`SELECT
			artist,
			song_name,
			song_text,
			chords,
			copyright
		FROM songs
		WHERE song_name = '%s';`,
		songName,
	)

	err = conn.QueryRow(context.Background(), query).Scan(
		&song.Artist,
		&song.SongName,
		&song.SongText,
		&song.Chords,
		&song.Copyright,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	return song
}

func stdSqlWay() {
	// continue from minute 11: https://www.youtube.com/watch?v=2XCaKYH0Ydo
	dsn := url.URL{ // "data source name": string of the url to the database
		Scheme: "postgres",
		Host:   "localhost:5432",
		User:   url.UserPassword("lyricsapi", "lyricsapi"),
		Path:   "lyricsapi",
	}
	db, err := sql.Open("pgx", dsn.String()) // returns a pool of connections
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open", err)
	}
	defer db.Close()
}
