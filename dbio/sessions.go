package dbio

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/davidkuda/lyricsapi/models"
)

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