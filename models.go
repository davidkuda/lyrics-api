package main

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	EMail     string `json:"email"`
	Password  string `json:"password"` // a hash of a password
	CreatedAt string `json:"-"`        // a hyphen means it's not put into the json
	UpdatedAt string `json:"-"`
}
