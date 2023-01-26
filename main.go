package main

import (
	"log"
	"net/http"
)

type Songs []Song

// Song contains all data related to a piece of music
type Song struct {
	// slug of the song, song name with hyphens, e.g. "wish-you-were-here"
	SongID string `json:"song_id"`
	// artist of the song, e.g. "Pink Floyd"
	Artist string `json:"artist"`
	// name of the song
	SongName string `json:"song_name"`
	// lyrics, text of the song
	SongText string `json:"song_text,omitempty"`
	// chords of the song, plain text
	Chords string `json:"chords,omitempty"`
	// copyright information of the song
	Copyright string `json:"copyright,omitempty"`
	// Covers: list of URLs to great covers, e.g. on YouTube
	Covers []string `json:"covers,omitempty"`
}

func main() {
	http.HandleFunc("/songs/", handleSongs)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
