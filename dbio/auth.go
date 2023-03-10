package dbio

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"

	"github.com/davidkuda/lyricsapi/config"
	"github.com/davidkuda/lyricsapi/internal/data"
	"github.com/davidkuda/lyricsapi/models"
)

func GetUserByEmail(email string, cfg config.AppConfig) (*models.User, error) {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		cfg.Logger.Printf("sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := `
	select
		id,
		email,
		password,
		created_at,
		updated_at
	from users
	where email = $1`

	var user models.User
	row := conn.QueryRowContext(context.Background(), query, email)

	if row.Err(); err != nil {
		cfg.Logger.Println("conn.QueryRow", err)
		return &user, errors.New("QueryNotSuccesful")
	}

	if err := row.Scan(
		&user.ID,
		&user.EMail,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		cfg.Logger.Println("row.Scan:", err)
		return nil, err
	}

	return &user, nil
}

func CreateNewUser(u *models.User, cfg config.AppConfig) error {
	// TODO: Add a salt
	// TODO: Check for length
	encrPW, err := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
	if err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := `
		INSERT INTO users (email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`

	res, err := conn.ExecContext(ctx, query, u.EMail, encrPW, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return err
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nRows != 1 {
		return fmt.Errorf("expected 1 row to be inserted, Got: %v", nRows)
	}

	return nil
}

// InsertToken() adds the data for a specific token to the tokens table.
func InsertToken(token *data.Token, cfg config.AppConfig) error {
	query := `
        INSERT INTO tokens (hash, email, expiry, scope) 
        VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserEMail, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := cfg.DB.ExecContext(ctx, query, args...)
	return err
}

// DeleteAllForUser() deletes all tokens for a specific user and scope.
func DeleteAllTokensForUser(scope string, userEmail string, cfg config.AppConfig) error {
	query := `
        DELETE FROM tokens 
        WHERE scope = $1 AND email = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := cfg.DB.ExecContext(ctx, query, scope, userEmail)
	return err
}

// retrieve the details of the user associated with a particular activation token
func GetForToken(tokenScope, tokenPlaintext string, cfg config.AppConfig) (*models.User, error) {
	// Calculate the SHA-256 hash of the plaintext token provided by the client.
	// Remember that this returns a byte *array* with length 32, not a slice.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT users.email, users.created_at, users.password
        FROM users
        INNER JOIN tokens
        ON users.email = tokens.email
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	// Create a slice containing the query arguments. Notice how we use the [:] operator
	// to get a slice containing the token hash, rather than passing in the array (which
	// is not supported by the pq driver), and that we pass the current time as the
	// value to check against the token expiry.
	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user models.User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query, scanning the return values into a User struct. If no matching
	// record is found we return an ErrRecordNotFound error.
	err := cfg.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.EMail,
		&user.CreatedAt,
		&user.Password,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("ErrRecordNotFound")
		default:
			return nil, err
		}
	}

	// Return the matching user.
	return &user, nil
}
