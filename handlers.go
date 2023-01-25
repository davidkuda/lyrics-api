package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// handler: get all songs
func handleListSongs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// get list of songs from db io
	songs := ListSongs()
	// ? how to only send the fields Song.Artist and Song.SongName? i.e. omit SongText
	body, err := json.Marshal(songs)
	if err != nil {
		status := http.StatusInternalServerError
		log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, err)
		http.Error(w, http.StatusText(status), status)
	
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)

}
