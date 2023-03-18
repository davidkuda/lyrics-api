package handlers

import (
	"net/http"

	"github.com/davidkuda/lyricsapi/config"

	"github.com/alexedwards/scs/v2"
)

type Application struct {
	Config         config.AppConfig
	Handler        func(w http.ResponseWriter, r *http.Request, config config.AppConfig)
	SessionManager *scs.SessionManager
}

func (app Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.Handler(w, r, app.Config)
}
