package dbio

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	
	"github.com/davidkuda/lyricsapi/lyricsapi"
)

func ListSongs() []string {
	return []string{}
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

	err = conn.QueryRow(context.Background(), "select artist from songs;").Scan(&song.Artist)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(song)
	return song
}
