package dbio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"

	"github.com/davidkuda/lyricsapi/config"
	"github.com/davidkuda/lyricsapi/internal/data"
	"github.com/davidkuda/lyricsapi/models"
)

var ErrSongDoesNotExist = errors.New("Song does not exist")

// returns a pool of connections to the postgres db according to the args
func GetDatabaseConn(dbAddr, dbName, dbUser, dbPassword string) (*sql.DB, error) {
	// "data source name": string of the url to the database
	dsn := url.URL{
		Scheme: "postgres",
		Host:   dbAddr,
		User:   url.UserPassword(dbUser, dbPassword),
		Path:   dbName,
	}
	return sql.Open("pgx", dsn.String())
}

func ListSongs(cfg config.AppConfig) models.Songs {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	query := "SELECT artist, song_name, song_id FROM songs ORDER BY artist ASC;"
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	var songs models.Songs
	var artist, song, song_id string
	for rows.Next() {
		rows.Scan(&artist, &song, &song_id)
		song := models.Song{Artist: artist, SongName: song, SongID: song_id}
		songs = append(songs, song)
	}

	return songs
}

func GetSong(songID string, cfg config.AppConfig) (models.Song, error) {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	song := models.Song{}

	query := `
		SELECT
			song_id,
			artist,
			song_name,
			song_text,
			chords,
			copyright
		FROM songs
		WHERE song_id = $1`

	row := conn.QueryRowContext(context.Background(), query, songID)

	if row.Err(); err != nil {
		cfg.Logger.Println("conn.QueryRow", err)
		return song, errors.New("QueryNotSuccesful")
	}

	row.Scan(
		&song.SongID,
		&song.Artist,
		&song.SongName,
		&song.SongText,
		&song.Chords,
		&song.Copyright,
	)

	if len(song.SongName) == 0 {
		return song, ErrSongDoesNotExist
	}

	return song, nil
}

func CreateSong(s *models.Song, cfg config.AppConfig) error {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := `
		INSERT INTO songs (
			song_id,
			artist,
			song_name,
			song_text,
			chords,
			copyright
		) VALUES ($1, $2, $3, $4, $5, $6);`

	if _, err := conn.ExecContext(
		ctx, query, s.SongID, s.Artist, s.SongName, s.SongText, s.Chords, s.Copyright,
	); err != nil {
		cfg.Logger.Println("conn.ExecContext:", err)
		return err
	}

	return nil
}

func DeleteSong(songID string, cfg config.AppConfig) error {
	ctx := context.Background()
	conn, err := cfg.DB.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	query := "DELETE FROM songs WHERE song_id = $1;"

	res, err := conn.ExecContext(ctx, query, songID)
	if err != nil {
		cfg.Logger.Println("conn.ExecContext:", err)
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		cfg.Logger.Println("res.RowsAffected:", err)
		return err
	}
	if n == 0 {
		return errors.New("Delete failed: Did not find song with id " + songID)
	}

	return nil
}

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

// Insert() adds the data for a specific token to the tokens table.
func Insert(token *data.Token, cfg config.AppConfig) error {
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
