package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/davidkuda/lyricsapi/dbio"
	"github.com/davidkuda/lyricsapi/models"
)

// /songs
func (a *Application) HandleSongsFixedPath(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		listSongs(w, r, a)
		return
	}

	if r.Method == http.MethodPost {
		// check if user is authenticated
		user := a.contextGetUser(r)
		if user.IsAnonymous() {
			a.authenticationRequiredResponse(w, r)
			return
		}
		a.createSong(w, r)
		return
	}

	if r.Method == http.MethodOptions {
		// CORS preflight request
		user := a.contextGetUser(r)
		if user.IsAnonymous() {
			a.authenticationRequiredResponse(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

// /songs/:id
func (a *Application) HandleSongsSubtreePath(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/songs/")

	if r.Method == http.MethodGet {
		returnSong(w, r, id, a)
		return
	}

	ok, _ := a.hasValidSessionCookie(w, r)
	if !ok {
		a.errorJSON(w, errors.New("HandleSong beyond Get: Invalid Session"), http.StatusUnauthorized)
	}

	if r.Method == http.MethodDelete {
		a.handleDeleteSong(w, id)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

func (app *Application) createSong(w http.ResponseWriter, r *http.Request) {
	s := models.Song{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		app.Logger.Println("io.ReadAll:", err)
	}
	defer r.Body.Close()
	if err := json.Unmarshal(data, &s); err != nil {
		app.Logger.Println("json.Unmarshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := dbio.CreateSong(&s, app.DB, app.Logger); err != nil {
		app.Logger.Println("dbio.CreateSong:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success: Created New Song"))
}

func (app *Application) handleDeleteSong(w http.ResponseWriter, songID string) {
	if err := dbio.DeleteSong(songID, app.DB, app.Logger); err != nil {
		app.Logger.Println("dbio.DeleteSong:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Success: Deleted Song with ID " + songID))
}

func listSongs(w http.ResponseWriter, r *http.Request, app *Application) {
	songs := dbio.ListSongs(app.DB)
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

func returnSong(w http.ResponseWriter, r *http.Request, id string, app *Application) {
	song, err := dbio.GetSong(id, app.DB, app.Logger)

	if err != nil {
		if err == dbio.ErrSongDoesNotExist {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			resp := make(map[string]string)
			resp["message"] = err.Error()
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				app.Logger.Printf("Error happened in JSON marshal. Err: %s", err)
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
