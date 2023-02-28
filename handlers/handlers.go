package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/davidkuda/lyricsapi/auth"
	"github.com/davidkuda/lyricsapi/config"
	"github.com/davidkuda/lyricsapi/dbio"
	"github.com/davidkuda/lyricsapi/models"
)

type Application struct {
	Config  config.AppConfig
	Handler func(w http.ResponseWriter, r *http.Request, config config.AppConfig)

	dbio DatabaseRepo

	Domain string

	Auth         auth.Auth
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
}

func (app Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.Handler(w, r, app.Config)
}

type DatabaseRepo interface {
	Connection() *sql.DB
	getDatabaseConn(dbAddr, dbName, dbUser, dbPassword string) (*sql.DB, error)
	ListSongs(cfg config.AppConfig) models.Songs
	GetSong(songName string, cfg config.AppConfig) (models.Song, error)
	GetUserByEmail(email string, cfg config.AppConfig) (*models.User, error)
}

// todo: log via middleware, not inside handlers
// ? how can you write logs to a file? can you write to stdout and to a file? (i.e. to multiple files?)
type requestLog struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	BodySize int64  `json:"content_length"`
	Protocol string `json:"protocol"`
}

func logRequest(r *http.Request, cfg *config.AppConfig) {
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
	cfg.Logger.Println(string(j))
}

func (app *Application) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	app.Config.Logger.Println("Handling HealthCheck Request")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (a *Application) HandleSongs(w http.ResponseWriter, r *http.Request) {
	logRequest(r, &a.Config)
	if r.Method == http.MethodGet {
		if len(r.URL.Path) > len("/songs/") {
			id := strings.TrimPrefix(r.URL.Path, "/songs/")
			returnSong(w, r, id, a.Config)
		} else {
			listSongs(w, r, a.Config)
		}

	} else if r.Method == http.MethodPost {
		_, _, err := a.Auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			// log error: do not return to the Client, but log internally for debugging.
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		a.handleCreateSong(w, r)

	} else if r.Method == http.MethodDelete {
		if _, _, err := a.Auth.GetTokenFromHeaderAndVerify(w, r); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		a.handleDeleteSong(w, r)

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (app *Application) handleCreateSong(w http.ResponseWriter, r *http.Request) {
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

func (app *Application) handleDeleteSong(w http.ResponseWriter, r *http.Request) {
	s := models.Song{}
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		app.Config.Logger.Println("io.ReadAll:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(data, &s); err != nil {
		app.Config.Logger.Println("json.Unmarshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := dbio.DeleteSong(s.SongID, app.Config); err != nil {
		app.Config.Logger.Println("dbio.DeleteSong:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success: Deleted Song with ID " + s.SongID))
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
