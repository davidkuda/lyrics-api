package dbio

import (
	"context"
	"errors"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"

	"github.com/davidkuda/lyricsapi/config"
	"github.com/davidkuda/lyricsapi/models"
)

func GetUserByName(name string, cfg config.AppConfig) (*models.User, error) {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		cfg.Logger.Printf("sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := `
	select
		name,
		password
	from users
	where name = $1`

	var user models.User
	row := conn.QueryRowContext(context.Background(), query, name)

	if row.Err(); err != nil {
		cfg.Logger.Println("conn.QueryRow", err)
		return &user, errors.New("QueryNotSuccesful")
	}

	if err := row.Scan(
		&user.Name,
		&user.Password,
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
		INSERT INTO users (name, password)
		VALUES ($1, $2)
	`

	res, err := conn.ExecContext(ctx, query, u.Name, encrPW)
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
