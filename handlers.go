package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// ? how can you write logs to a file? can you write to stdout and to a file? (i.e. to multiple files?)
type requestLog struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	BodySize int64  `json:"content_length"`
	Protocol string `json:"protocol"`
}

func logRequest(r *http.Request, cfg *appConfig) {
	l := requestLog{
		URL:      r.URL.String(),
		Method:   r.Method,
		BodySize: r.ContentLength,
		Protocol: r.Proto,
	}

	j, err := json.Marshal(&l)
	if err != nil {
		panic(err)
	}
	cfg.logger.Println(string(j))
}

func setupHandlers(mux *http.ServeMux, config appConfig) {
	mux.Handle("/songs/", &app{config: config, handler: handleSongs})
}

func handleSongs(w http.ResponseWriter, r *http.Request, cfg appConfig) {
	logRequest(r, &cfg)
	if len(r.URL.Path) > len("/songs/") {
		id := strings.TrimPrefix(r.URL.Path, "/songs/")
		returnSong(w, r, id, &cfg)
	} else {
		listSongs(w, r, &cfg)
	}
}

func listSongs(w http.ResponseWriter, r *http.Request, cfg *appConfig) {
	songs := ListSongs(cfg)
	// ? how to only send the fields Song.Artist and Song.SongName? i.e. omit SongText
	body, err := json.Marshal(songs)
	if err != nil {
		status := http.StatusInternalServerError
		log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, err)
		http.Error(w, http.StatusText(status), status)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func returnSong(w http.ResponseWriter, r *http.Request, id string, cfg *appConfig) {
	// TODO: Validate if song in songs; maybe in dbio? or here? dbio could return err if song not in db
	// song := GetSong(id)
	song := stdSqlWay()
	body, err := json.Marshal(song)
	if err != nil {
		status := http.StatusInternalServerError
		log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, err)
		http.Error(w, http.StatusText(status), status)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
