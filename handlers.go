package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/davidkuda/lyricsapi/config"
	"github.com/davidkuda/lyricsapi/dbio"
)

type application struct {
	config  config.AppConfig
	handler func(w http.ResponseWriter, r *http.Request, config config.AppConfig)

	dbio DatabaseRepo

	Domain string

	auth         Auth
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
}

func (app application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.handler(w, r, app.config)
}

type DatabaseRepo interface {
	Connection() *sql.DB
	getDatabaseConn(dbAddr, dbName, dbUser, dbPassword string) (*sql.DB, error)
	ListSongs(cfg appConfig) Songs
	GetSong(songName string, cfg appConfig) (Song, error)
	GetUserByEmail(email string, cfg appConfig) (*User, error)
}

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
	app.config.Logger.Println("Handling HealthCheck Request")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (a *application) handleSongs(w http.ResponseWriter, r *http.Request) {
	logRequest(r, &a.config)
	if r.Method == http.MethodGet {
		if len(r.URL.Path) > len("/songs/") {
			id := strings.TrimPrefix(r.URL.Path, "/songs/")
			returnSong(w, r, id, a.config)
		} else {
			listSongs(w, r, a.config)
		}
	} else if r.Method == http.MethodPost {
		_, _, err := a.auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		a.handleCreateSong(w, r)
		// write new song to db
	} else if r.Method == http.MethodDelete {
		_, _, err := a.auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	} else {
		listSongs(w, r, a.config)
	}
}

func (app *application) handleCreateSong(w http.ResponseWriter, r *http.Request) {
	s := Song{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		app.config.Logger.Println("io.ReadAll:", err)
	}
	defer r.Body.Close()
	if err := json.Unmarshal(data, &s); err != nil {
		app.config.Logger.Println("json.Unmarshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := dbio.CreateSong(&s, app.config); err != nil {
		app.config.Logger.Println("dbio.CreateSong:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success: Created New Song"))
}

func (app *application) handleDeleteSong(w http.ResponseWriter, r *http.Request) {
	s := Song{}
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		app.config.Logger.Println("io.ReadAll:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(data, &s); err != nil {
		app.config.Logger.Println("json.Unmarshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := dbio.DeleteSong(s.SongID, app.config); err != nil {
		app.config.Logger.Println("dbio.DeleteSong:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success: Deleted Song with ID " + s.SongID))
}
func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	logRequest(r, &app.config)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		status := http.StatusInternalServerError
		log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, err)
		http.Error(w, http.StatusText(status), status)
	}

	t := time.Now()
	newUser := User{CreatedAt: t, UpdatedAt: t}
	// unmarshal E-Mail and Password from payload of the request
	json.Unmarshal(data, &newUser)

	// TODO: check if password is hashable, pw + salt should not exceed max length of bcrypt
	if err := dbio.CreateNewUser(&newUser, app.config); err != nil {
		status := http.StatusInternalServerError
		log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success: User Created"))
	return
}

func (app *application) signin(w http.ResponseWriter, r *http.Request) {
	logRequest(r, &app.config)

	// read json payload
	var requestPayload struct {
		Email    string `string:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
	}

	// validate user against database
	user, err := dbio.GetUserByEmail(requestPayload.Email, app.config)
	if err != nil {
		app.errorJSON(w, errors.New("GetUserByEmail: failed"), http.StatusBadRequest)
		return
	}

	// check password in db
	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// create a jwt user
	u := JWTUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		// return json error
		return
	}

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)

	app.writeJSON(w, http.StatusAccepted, tokens)
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
