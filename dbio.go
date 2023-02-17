package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

var ErrSongDoesNotExist = errors.New("Song does not exist")

type DatabaseRepo interface {
	Connection() *sql.DB
	getDatabaseConn(dbAddr, dbName, dbUser, dbPassword string) (*sql.DB, error)
	ListSongs(cfg appConfig) Songs
	GetSong(songName string, cfg appConfig) (Song, error)
	GetUserByEmail(email string, cfg appConfig) (*User, error)
}

// returns a pool of connections to the postgres db according to the args
func getDatabaseConn(dbAddr, dbName, dbUser, dbPassword string) (*sql.DB, error) {
	// "data source name": string of the url to the database
	dsn := url.URL{
		Scheme: "postgres",
		Host:   dbAddr,
		User:   url.UserPassword(dbUser, dbPassword),
		Path:   dbName,
	}
	return sql.Open("pgx", dsn.String())
}

func ListSongs(cfg appConfig) Songs {
	ctx := context.Background()
	conn, err := cfg.db.Conn(ctx)
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

	var songs Songs
	var artist, song, song_id string
	for rows.Next() {
		rows.Scan(&artist, &song, &song_id)
		song := Song{Artist: artist, SongName: song, SongID: song_id}
		songs = append(songs, song)
	}

	return songs
}

func GetSong(songName string, cfg appConfig) (Song, error) {
	ctx := context.Background()
	conn, err := cfg.db.Conn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
	}
	defer conn.Close()

	song := Song{}

	query := `
		SELECT
			song_id,
			artist,
			song_name,
			song_text,
			chords,
			copyright
		FROM songs
		WHERE song_name = $1`

	row := conn.QueryRowContext(context.Background(), query, songName)

	if row.Err(); err != nil {
		cfg.logger.Println("conn.QueryRow", err)
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

func GetUserByEmail(email string, cfg appConfig) (*User, error) {
	ctx := context.Background()
	conn, err := cfg.db.Conn(ctx)
	if err != nil {
		cfg.logger.Printf("sql.Open: %v\n", err)
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

	var user User
	row := conn.QueryRowContext(context.Background(), query, email)

	if row.Err(); err != nil {
		cfg.logger.Println("conn.QueryRow", err)
		return &user, errors.New("QueryNotSuccesful")
	}

	if err := row.Scan(
		&user.ID,
		&user.EMail,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		cfg.logger.Println("row.Scan:", err)
		return nil, err
	}

	return &user, nil
}

func CreateNewUser(u *User, cfg appConfig) error {
	// TODO: Add a salt
	// TODO: Check for length
	encrPW, err := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
	if err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := cfg.db.Conn(ctx)
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
