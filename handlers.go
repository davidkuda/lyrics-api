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

func (app *application) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	app.config.logger.Println("Handling HealthCheck Request")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (a *application) handleSongs(w http.ResponseWriter, r *http.Request) {
	logRequest(r, &a.config)
	if len(r.URL.Path) > len("/songs/") {
		id := strings.TrimPrefix(r.URL.Path, "/songs/")
		returnSong(w, r, id, a.config)
	} else {
		listSongs(w, r, a.config)
	}
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	// read json payload

	// validate user against database

	// check password in db

	// create a jwt user

	u := JWTUser{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
	}

	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		// return json error
		return
	}
	app.config.logger.Println(tokens.Token)
	w.Write([]byte(tokens.Token))
}

func listSongs(w http.ResponseWriter, r *http.Request, cfg appConfig) {
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

func returnSong(w http.ResponseWriter, r *http.Request, id string, cfg appConfig) {
	song, err := GetSong(id, cfg)

	if err != nil {
		if err == ErrSongDoesNotExist {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			resp := make(map[string]string)
			resp["message"] = err.Error()
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				cfg.logger.Printf("Error happened in JSON marshal. Err: %s", err)
			}
			w.Write(jsonResp)
			return
		}
	}

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
