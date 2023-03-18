package main

import (
	"net/http"

	"github.com/davidkuda/lyricsapi/handlers"
)

// Go’s servemux supports two different types of URL patterns:
// fixed paths and subtree paths. Fixed paths do not end with a
// trailing slash, whereas subtree paths do end with a trailing
// slash.

func setupHandlers(mux *http.ServeMux, app handlers.Application) {
	mux.HandleFunc("/healthz", app.HandleHealthCheck)
	mux.HandleFunc("/songs", app.HandleSongsFixedPath)
	mux.HandleFunc("/songs/", app.HandleSongsSubtreePath)
	mux.HandleFunc("/v1/tokens/authentication", app.CreateAuthenticationTokenHandler)
}
