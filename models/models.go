package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Songs []Song

// Song contains all data related to a piece of music
// ID: slug of the song, song name with hyphens, e.g. "wish-you-were-here"
// Artist: artist of the song, e.g. "Pink Floyd"
// Name: name of the song
// Text: lyrics, text of the song
// Chords: chords of the song, plain text
// Copyright: copyright information of the song
// Covers: list of URLs to great covers, e.g. on YouTube
type Song struct {
	ID        string   `json:"id"`
	Artist    string   `json:"artist"`
	Name      string   `json:"name"`
	Text      string   `json:"lyrics,omitempty"`
	Chords    string   `json:"chords,omitempty"`
	Copyright string   `json:"copyright,omitempty"`
	Covers    []string `json:"covers,omitempty"`
}

type User struct {
	Name      string    `json:"name"`
	Password  string    `json:"password"` // a hash of a password
	CreatedAt time.Time `json:"-"`        // a hyphen means it's not put into the json
	UpdatedAt time.Time `json:"-"`
}

var AnonymousUser = &User{}

// Check if a User instance is the AnonymousUser.
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (u *User) PasswordMatches(plainText string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// invalid password
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
