package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func handleSongs(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) > len("/songs/") {
		id := strings.TrimPrefix(r.URL.Path, "/songs/")
		returnSong(w, r, id)
	} else {
		listSongs(w, r)
	}
}

func listSongs(w http.ResponseWriter, r *http.Request) {
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

func returnSong(w http.ResponseWriter, r *http.Request, id string) {
	// TODO: Validate if song in songs; maybe in dbio? or here? dbio could return err if song not in db
	song := GetSong(id)
	body, err := json.Marshal(song)
	if err != nil {
		status := http.StatusInternalServerError
		log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, err)
		http.Error(w, http.StatusText(status), status)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
