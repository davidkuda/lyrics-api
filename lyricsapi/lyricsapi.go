package lyricsapi

import (
	"net/http"
)

type Songs []Song

type Song struct {
	Artist    string
	SongName  string
	SongText  string
	Chords    string
	Copyright string
	Covers    []string // list of URLs to great covers, e.g. on YouTube
}

// handler: get all songs
func handleListSongs(w http.ResponseWriter, r *http.Request) {
	// get list of songs from db io
}
