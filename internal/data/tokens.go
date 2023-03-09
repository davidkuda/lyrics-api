package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
    Plaintext string    `json:"token"`
    Hash      []byte    `json:"-"`
    UserID    int64     `json:"-"`
    Expiry    time.Time `json:"expiry"`
    Scope     string    `json:"-"`
}

func GenerateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	// sha256.Sum256() function returns an *array* of length 32, so to make it easier to
	// work with we convert it to a slice using the [:] operator before storing it.
	token.Hash = hash[:]

	return token, nil
}
