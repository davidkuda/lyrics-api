package handlers

import (
	"database/sql"
	"log"

	"github.com/alexedwards/scs/v2"
)

type Application struct {
	// Handler        func(w http.ResponseWriter, r *http.Request, config config.AppConfig)
	Logger *log.Logger
	DB     *sql.DB
	CORS   struct {
		TrustedOrigins []string
	}

	SessionManager *scs.SessionManager
}

// func (app Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	app.Handler(w, r)
// }
