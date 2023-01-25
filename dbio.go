package dbio

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"

	"github.com/davidkuda/lyricsapi/lyricsapi"
)

func ListSongs() []string {
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

	var songs []string
	var artist, song string
	for rows.Next() {
		rows.Scan(&artist, &song)
		songs = append(songs, fmt.Sprintf("%s -- %s", artist, song))
	}

	return songs
}

func GetSong(songName string) lyricsapi.Song {
	// TODO: Validate input, avoid SQLInjection, check against all available songs, store all songs in memory for fast check
	os.Setenv("DATABASE_URL", "postgres://lyricsapi:lyricsapi@localhost:5432/lyricsapi")
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	song := lyricsapi.Song{}

	query := fmt.Sprintf("SELECT artist, song_name, song_text FROM songs WHERE song_name = '%s';", songName)
	err = conn.QueryRow(context.Background(), query).Scan(&song.Artist, &song.SongName, &song.SongText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	return song
}
