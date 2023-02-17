package main

import "net/http"

// /v1/lyrics/ GET -> return a list of all available song names
// /v1/lyrics/ POST -> create a new song
// /v1/lyrics/{song-name} PATCH -> update an existing song
// PATCH vs PUT: Patch for elements, PUT for replacing the whole row
// /v1/lyrics/{song-name} GET -> return songtext with metadata
func setupHandlers(mux *http.ServeMux, app application) {
	mux.HandleFunc("/songs", app.handleSongs)
	mux.HandleFunc("/healthz", app.handleHealthCheck)
	mux.HandleFunc("/signup", app.signup)
	mux.HandleFunc("/signin", app.signin)
}
