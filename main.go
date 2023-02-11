package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"
)

type application struct {
	config appConfig

	Domain string

	auth         Auth
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
}

type appConfig struct {
	logger *log.Logger
	db     *sql.DB
}

type Songs []Song

// Song contains all data related to a piece of music
// SongID: slug of the song, song name with hyphens, e.g. "wish-you-were-here"
// Artist: artist of the song, e.g. "Pink Floyd"
// SongName: name of the song
// SongText: lyrics, text of the song
// Chords: chords of the song, plain text
// Copyright: copyright information of the song
// Covers: list of URLs to great covers, e.g. on YouTube
type Song struct {
	SongID    string   `json:"song_id"`
	Artist    string   `json:"artist"`
	SongName  string   `json:"song_name"`
	SongText  string   `json:"song_text,omitempty"`
	Chords    string   `json:"chords,omitempty"`
	Copyright string   `json:"copyright,omitempty"`
	Covers    []string `json:"covers,omitempty"`
}

// in main, it's ok to log.Fatal or to os.Exit(1), but not in other places
func main() {
	var app application

	app.JWTSecret = os.Getenv("JWT_SECRET")
	app.JWTIssuer = os.Getenv("JWT_ISSUER")
	app.JWTAudience = os.Getenv("JWT_AUDIENCE")
	app.CookieDomain = os.Getenv("COOKIE_DOMAIN")

	app.auth = Auth{
		Issuer:        app.JWTIssuer,
		Audience:      app.JWTAudience,
		Secret:        app.JWTSecret,
		TokenExpiry:   time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		Cookie: Cookie{
			Path:   "/",
			Name:   "__Host-refresh_token",
			Domain: app.CookieDomain,
		},
	}

	dbAddr := os.Getenv("DB_ADDR")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	// ! alternative: if dbAddr == "" -- shorter and easier to read
	if len(dbAddr) == 0 || len(dbName) == 0 || len(dbUser) == 0 || len(dbPassword) == 0 {
		log.Fatal("Must specify db details as env var: DB_ADDR, DB_NAME, DB_USER and DB_PASSWORD")
	}

	db, err := getDatabaseConn(dbAddr, dbName, dbUser, dbPassword)
	if err != nil {
		log.Fatalf("getDatabaseConn(): %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("db.Ping(): %v", err)
	}

	app.config = appConfig{
		// ? how to append log to a file or to a database? use a Tee on os level; Stdout and Stderr is the conventional choice.
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
		db:     db,
	}

	mux := http.NewServeMux()
	setupHandlers(mux, app)

	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		listenAddr = ":8008"
	}

	log.Printf("Starting app; listening on port %s", listenAddr)
	// ? ListenAndServe: If you terminate the process, the last requests may get lost. Check Ardan Labs "Service" to see an alternative.
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
