package models

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/base32"
    "time"
)

type Token struct {
    Plaintext string
    Hash      []byte
    UserName  string
    Expiry    time.Time
}

func GenerateToken(userName string, ttl time.Duration) (*Token, error) {
    // Notice that we add the provided ttl (time-to-live) duration parameter to the 
    // current time to get the expiry time
    token := Token{
        UserName: userName,
        Expiry: time.Now().Add(ttl),
    }

    // Initialize a zero-valued byte slice with a length of 16 bytes.
    randomBytes := make([]byte, 16)

    // Use the Read() function from the crypto/rand package to fill the byte slice with 
    // random bytes from your operating system's CSPRNG. This will return an error if 
    // the CSPRNG fails to function correctly.
    _, err := rand.Read(randomBytes)
    if err != nil {
        return nil, err
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
    token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

    // Generate a SHA-256 hash of the plaintext token string. This will be the value 
    // that we store in the `hash` field of our database table. Note that the 
    // sha256.Sum256() function returns an *array* of length 32, so to make it easier to  
    // work with we convert it to a slice using the [:] operator before storing it.
    hash := sha256.Sum256([]byte(token.Plaintext))
    token.Hash = hash[:]

    return &token, nil
}
