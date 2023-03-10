package handlers

import (
	"net/http"

	"github.com/davidkuda/lyricsapi/config"
)

type Application struct {
	Config  config.AppConfig
	Handler func(w http.ResponseWriter, r *http.Request, config config.AppConfig)
}

func (app Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.Handler(w, r, app.Config)
}
