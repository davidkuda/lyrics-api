package main

import (
	"net/http"

	"github.com/davidkuda/lyricsapi/dbio"
)

// handler: get all songs
func HandleListSongs(w http.ResponseWriter, r *http.Request) {
	// get list of songs from db io
	songs := dbio.ListSongs()
}
