package dbio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/davidkuda/lyricsapi/models"
)

var ErrNoTokenFound = errors.New("NoTokenFound")

func CreateNewSession(t models.SessionToken, db *sql.DB) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := `
		INSERT INTO sessions (token, user_name, expiry)
		VALUES ($1, $2, $3)
	`

	res, err := conn.ExecContext(ctx, query, t.Token, t.UserName, t.Expiry)
	if err != nil {
		return fmt.Errorf("conn.ExecContext: %v\n", err)
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

func GetSessionToken(token string, db *sql.DB) (models.SessionToken, error) {
	t := models.SessionToken{}

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return t, fmt.Errorf("sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := `
		SELECT token, user_name, expiry
		FROM sessions
		WHERE token = '$1';
	`

	rows, err := conn.QueryContext(ctx, query, token)
	if err != nil {
		return t, fmt.Errorf("conn.QueryContext: %v\n", err)
	}

	if err = rows.Scan(&t.Token, &t.UserName, &t.Expiry); err != nil {
		return t, fmt.Errorf("rows.Scan: %v\n", err)
	}

	if t.Token == "" {
		return t, ErrNoTokenFound
	}

	return t, nil
}

func DeleteExpiredTokens(db *sql.DB) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := `
		DELETE FROM sessions
		WHERE expiry < NOW();
	`

	_, err = conn.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("conn.ExecContext: %v\n", err)
	}
	
	return nil
}
