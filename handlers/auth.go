package handlers

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/davidkuda/lyricsapi/dbio"
	"github.com/davidkuda/lyricsapi/models"
)

func (app *Application) Authenticate(w http.ResponseWriter, r *http.Request) {
	// Parse the userName and password from the request body.
	var input struct {
		UserName string `json:"userName"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// TODO: validate email and password

	user, err := dbio.GetUserByName(input.UserName, app.DB, app.Logger)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Check if the provided password matches the actual password for the user.
	match, err := user.PasswordMatches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// If the passwords don't match, then we call the app.invalidCredentialsResponse()
	// helper again and return.
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// create session token
	t := models.SessionToken{}
	t.UserName = input.UserName

	ttl := 24 * time.Hour // ttl == time to live
	t.Expiry = time.Now().Add(ttl)

	token, err := generateToken()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	t.Token = token

	if err := dbio.CreateNewSession(t, app.DB); err != nil {
		app.Logger.Println(err)
	}

	// send session token as cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    t.Token,
		Expires:  t.Expiry,
		Secure:   true,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusCreated)
}

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

	newUser := models.User{}

	// unmarshal E-Mail and Password from payload of the request
	json.Unmarshal(data, &newUser)

	// TODO: check if password is hashable, pw + salt should not exceed max length of bcrypt
	if err := dbio.CreateNewUser(&newUser, app.DB, app.Logger); err != nil {
		status := http.StatusInternalServerError
		log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success: User Created"))
}

func generateToken() (string, error) {
	// Initialize a zero-valued byte slice with a length of 16 bytes.
	randomBytes := make([]byte, 16)

	// Use the Read() function from the crypto/rand package to fill the byte slice with
	// random bytes from your operating system's CSPRNG. This will return an error if
	// the CSPRNG fails to function correctly.
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the byte slice to a base-32-encoded string and assign it to the token
	// Plaintext field. This will be the token string that we send to the user in their
	// welcome email. They will look similar to this:
	//
	// Y3QMGX3PJ3WLRL2YRTQGQ6KRHU
	//
	// Note that by default base-32 strings may be padded at the end with the =
	// character. We don't need this padding character for the purpose of our tokens, so
	// we use the WithPadding(base32.NoPadding) method in the line below to omit them.
	tokenPlaintext := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	return tokenPlaintext, nil
}

func (app *Application) HasActiveSession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session")
	if err != nil {
		if err == http.ErrNoCookie {
			app.errorJSON(w, errors.New("Not Authenticated"), http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientToken := c.Value

	t, err := dbio.GetSessionToken(clientToken, app.DB)
	if err != nil {
		if err == dbio.ErrNoTokenFound {
			app.errorJSON(w, errors.New("Invalid Session Token"), http.StatusUnauthorized)
			return
		}
	}

	if t.Expiry.Before(time.Now()) {
		app.errorJSON(w, errors.New("Session Expired"), http.StatusUnauthorized)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"session": t.UserName}, nil)
}
