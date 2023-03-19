package handlers

import (
	"database/sql"
	"log"
)

type Application struct {
	// Handler        func(w http.ResponseWriter, r *http.Request, config config.AppConfig)
	Logger *log.Logger
	DB     *sql.DB
	CORS   struct {
		TrustedOrigins []string
	}
}

// func (app Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	app.Handler(w, r)
// }
