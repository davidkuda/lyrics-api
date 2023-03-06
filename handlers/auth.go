package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/davidkuda/lyricsapi/auth"
	"github.com/davidkuda/lyricsapi/dbio"
	"github.com/davidkuda/lyricsapi/models"
)

func (app *Application) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		status := http.StatusInternalServerError
		log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, err)
		http.Error(w, http.StatusText(status), status)
	}

	t := time.Now()
	newUser := models.User{CreatedAt: t, UpdatedAt: t}
	// unmarshal E-Mail and Password from payload of the request
	json.Unmarshal(data, &newUser)

	// TODO: check if password is hashable, pw + salt should not exceed max length of bcrypt
	if err := dbio.CreateNewUser(&newUser, app.Config); err != nil {
		status := http.StatusInternalServerError
		log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success: User Created"))
	return
}

func (app *Application) Signin(w http.ResponseWriter, r *http.Request) {
	// origin := "http://localhost:3001"
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// if r.Method == "OPTIONS" {
	// 	w.Header().Set("Access-Control-Allow-Credentials", "true")
	// 	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
	// }

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
	user, err := dbio.GetUserByEmail(requestPayload.Email, app.Config)
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
	u := auth.JWTUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	tokens, err := app.Auth.GenerateTokenPair(&u)
	if err != nil {
		// return json error
		return
	}

	c := http.Cookie{
		Name:     "jwt",
		Value:    tokens.Token,
		MaxAge:   15 * 60,
		HttpOnly: true,
	}

	http.SetCookie(w, &c)

	refreshCookie := app.Auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)

	app.writeJSON(w, http.StatusAccepted, tokens)
}
