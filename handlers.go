package main

import (
	"net/http"
)

// handler: get all songs
func handleListSongs(w http.ResponseWriter, r *http.Request) {
	// get list of songs from db io
	songs := ListSongs()
	
}
