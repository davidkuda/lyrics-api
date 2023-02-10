package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var api_key = "much secret"
var SECRET = os.Getenv("JWT_SECRET")

func createJWT() string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	tokenStr, err := token.SignedString(SECRET)
	if err != nil {
		log.Fatal("createJWT: Failed to parse token to string")
	}

	return tokenStr
}

func validateJWT(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header["Token"] != nil {
				token, err := jwt.Parse(
					r.Header["Token"][0], func(t *jwt.Token) (interface{}, error) {
						_, ok := t.Method.(*jwt.SigningMethodHMAC)
						if !ok {
							w.WriteHeader(http.StatusUnauthorized)
							w.Write([]byte("not authorized"))
						}
						return SECRET, nil
					},
				)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("not authorized: " + err.Error()))
				}

				if token.Valid {
					next(w, r)
				}
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("not authorized"))
			}
		},
	)
}

func getJWT(w http.ResponseWriter, r *http.Request) {
	if r.Header["api_key"] == nil {
		return
	}
	if r.Header["api_key"][0] == api_key {
		token := createJWT()
		fmt.Fprintf(w, token)
	}
}
