package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/davidkuda/lyricsapi/config"
	"github.com/davidkuda/lyricsapi/dbio"
	"github.com/davidkuda/lyricsapi/models"
)

// /songs
func (a *Application) HandleSongsFixedPath(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		listSongs(w, r, a.Config)

	} else if r.Method == http.MethodPost {
		// check if user is authenticated
		user := a.contextGetUser(r)
		if user.IsAnonymous() {
			a.authenticationRequiredResponse(w, r)
			return
		}
		a.createSong(w, r)

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// /songs/:id
func (a *Application) HandleSongsSubtreePath(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/songs/")

	if r.Method == http.MethodGet {
		returnSong(w, r, id, a.Config)

	} else if r.Method == http.MethodDelete {
		// check if user is authenticated
		user := a.contextGetUser(r)
		if user.IsAnonymous() {
			a.authenticationRequiredResponse(w, r)
			return
		}
		a.handleDeleteSong(w, id)

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (app *Application) createSong(w http.ResponseWriter, r *http.Request) {
	s := models.Song{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		app.Config.Logger.Println("io.ReadAll:", err)
	}
	defer r.Body.Close()
	if err := json.Unmarshal(data, &s); err != nil {
		app.Config.Logger.Println("json.Unmarshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := dbio.CreateSong(&s, app.Config); err != nil {
		app.Config.Logger.Println("dbio.CreateSong:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success: Created New Song"))
}

func (app *Application) handleDeleteSong(w http.ResponseWriter, songID string) {
	if err := dbio.DeleteSong(songID, app.Config); err != nil {
		app.Config.Logger.Println("dbio.DeleteSong:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Success: Deleted Song with ID " + songID))
}

func listSongs(w http.ResponseWriter, r *http.Request, cfg config.AppConfig) {
	songs := dbio.ListSongs(cfg)
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

func returnSong(w http.ResponseWriter, r *http.Request, id string, cfg config.AppConfig) {
	song, err := dbio.GetSong(id, cfg)

	if err != nil {
		if err == dbio.ErrSongDoesNotExist {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			resp := make(map[string]string)
			resp["message"] = err.Error()
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				cfg.Logger.Printf("Error happened in JSON marshal. Err: %s", err)
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
