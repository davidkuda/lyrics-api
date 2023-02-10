package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Auth struct {
	Issuer        string
	Audience      string
	Secret        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	Cookie
}

type Cookie struct {
	Domain string
	Path   string
	Name   string
}

type JWTUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
}

func (j *Auth) GenerateTokenPair(user *JWTUser) (TokenPairs, error) {
	// Create a Token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	// sub == subject
	claims["sub"] = fmt.Sprint(user.ID)
	// aud == audience
	claims["aud"] = j.Audience
	// iss == issuer
	claims["iss"] = j.Issuer
	// iat == issued at
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"

	// Set the expiry for the JWT
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()
	
	// Create a signed token
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// Create a refresh token and set claims
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["sub"] = fmt.Sprint(user.ID)
	rtClaims["iat"] = time.Now().UTC().Unix()

	// Set the expiry for the refresh token
	rtClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()
	
	// Create signed refresh toke
	signedRefreshToken, err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}
	
	// Create TokenPairs and populate with signed tokens
	tokenPairs := TokenPairs{
		Token: signedAccessToken,
		RefreshToken: signedRefreshToken,
	}
	
	// Return TokenPairs
	return tokenPairs, nil
}
