package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
)

type appConfig struct {
	logger *log.Logger
	db     *sql.DB
}

type app struct {
	config appConfig
	handler func(w http.ResponseWriter, r *http.Request, config appConfig)
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handler(w, r, a.config)
}

type Songs []Song

// Song contains all data related to a piece of music
type Song struct {
	// slug of the song, song name with hyphens, e.g. "wish-you-were-here"
	SongID string `json:"song_id"`
	// artist of the song, e.g. "Pink Floyd"
	Artist string `json:"artist"`
	// name of the song
	SongName string `json:"song_name"`
	// lyrics, text of the song
	SongText string `json:"song_text,omitempty"`
	// chords of the song, plain text
	Chords string `json:"chords,omitempty"`
	// copyright information of the song
	Copyright string `json:"copyright,omitempty"`
	// Covers: list of URLs to great covers, e.g. on YouTube
	Covers []string `json:"covers,omitempty"`
}

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		listenAddr = ":8008"
	}

	dbAddr := os.Getenv("DB_ADDR")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	if len(dbAddr) == 0 || len(dbName) == 0 || len(dbUser) == 0 || len(dbPassword) == 0 {
		log.Fatal("Must specify db details as env var: DB_ADDR, DB_NAME, DB_USER and DB_PASSWORD")
	}

	db, err := getDatabaseConn(dbAddr, dbName, dbUser, dbPassword)
	if err != nil {
		log.Fatal("getDatabaseConn(): Unable to get conn")
	}

	cfg := appConfig{
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
		db:     db,
	}
	
	mux := http.NewServeMux()
	setupHandlers(mux, cfg)

	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
