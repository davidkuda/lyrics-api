package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/davidkuda/lyricsapi/dbio"
	"github.com/davidkuda/lyricsapi/handlers"
)

// in main, it's ok to log.Fatal or to os.Exit(1), but not in other places
func main() {
	var app handlers.Application

	app.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	dbAddr := os.Getenv("DB_ADDR")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	// ! alternative: if dbAddr == "" -- shorter and easier to read
	if len(dbAddr) == 0 || len(dbName) == 0 || len(dbUser) == 0 || len(dbPassword) == 0 {
		log.Fatal("Must specify db details as env var: DB_ADDR, DB_NAME, DB_USER and DB_PASSWORD")
	}

	db, err := dbio.GetDatabaseConn(dbAddr, dbName, dbUser, dbPassword)
	if err != nil {
		log.Fatalf("getDatabaseConn(): %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("db.Ping(): %v", err)
	}
	log.Printf("Connection to database established: %s@%s", dbUser, dbName)

	app.DB = db

	// list allowed cors origins separated by space
	allowedCORSOrigins := strings.Split(os.Getenv("ALLOWED_CORS_ORIGINS"), " ")
	app.CORS = struct{ TrustedOrigins []string }{allowedCORSOrigins}

	mux := http.NewServeMux()
	setupHandlers(mux, app)

	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		listenAddr = ":8032"
	}

	log.Printf("Starting app; listening on port %s", listenAddr)
	log.Fatal(http.ListenAndServe(
		listenAddr,
		app.LogRequests(
			app.EnableCORS(
				mux,
			),
		),
	))
}
