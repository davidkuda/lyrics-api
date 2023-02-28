package main

import (
	"net/http"

	"github.com/davidkuda/lyricsapi/handlers"
)

// /v1/lyrics/ GET -> return a list of all available song names
// /v1/lyrics/ POST -> create a new song
// /v1/lyrics/{song-name} PATCH -> update an existing song
// PATCH vs PUT: Patch for elements, PUT for replacing the whole row
// /v1/lyrics/{song-name} GET -> return songtext with metadata
func setupHandlers(mux *http.ServeMux, app handlers.Application) {
	mux.HandleFunc("/songs", app.HandleSongs)
	mux.HandleFunc("/healthz", app.HandleHealthCheck)
	mux.HandleFunc("/signup", app.Signup)
	mux.HandleFunc("/signin", app.Signin)
}
