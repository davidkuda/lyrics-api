package handlers

import (
	"net/http"
	"time"

	"github.com/davidkuda/lyricsapi/dbio"
	"github.com/davidkuda/lyricsapi/internal/data"
)

func (app *Application) CreateAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the email and password from the request body.
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// TODO: validate email and password

	user, err := dbio.GetUserByEmail(input.Email, app.Config)
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

	// Otherwise, if the password is correct, we generate a new token with a 24-hour
	// expiry time and the scope 'authentication'.
	token, err := data.GenerateToken(user.EMail, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	dbio.InsertToken(token, app.Config)

	// Encode the token to JSON and send it in the response along with a 201 Created
	// status code.
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
