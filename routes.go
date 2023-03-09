package main

import (
	"net/http"

	"github.com/davidkuda/lyricsapi/handlers"
)

func setupHandlers(mux *http.ServeMux, app handlers.Application) {
	mux.HandleFunc("/songs/", app.HandleSongs)
	mux.HandleFunc("/healthz", app.HandleHealthCheck)
	mux.HandleFunc("/signup", app.Signup)
	mux.HandleFunc("/signin", app.Signin)
}
