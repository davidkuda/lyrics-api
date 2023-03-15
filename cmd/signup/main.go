package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/davidkuda/lyricsapi/dbio"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	conn := DBConn()
	defer conn.Close()

	email := flag.String("email", "", "The email address of the new user")
	password := flag.String("password", "", "The password of the new user")
	deleteUser := flag.Bool("delete-user", false, "bool: whether user with given email address should be deleted")
	flag.Parse()

	if len(*email) == 0 {
		log.Fatal("Make sure to pass an email address")
	}

	if *deleteUser {
		delete(*email, conn)
	} else {
		create(*email, *password, conn)
	}
}

func DBConn() *sql.Conn {
	dbAddr := os.Getenv("DB_ADDR")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	db, err := dbio.GetDatabaseConn(dbAddr, dbName, dbUser, dbPassword)
	if err != nil {
		log.Fatalf("getDatabaseConn(): %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("db.Ping(): %v", err)
	}
	log.Printf("Connection to database established: %s@%s", dbUser, dbName)

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func delete(email string, conn *sql.Conn) {
	query := `
			DELETE FROM users
			WHERE email = $1
		`

	ctx := context.Background()
	res, err := conn.ExecContext(ctx, query, email)
	if err != nil {
		log.Fatal("conn.ExecContext: ", err)
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if nRows != 1 {
		log.Fatalf("expected 1 row to be deleted, Got: %v", nRows)
	}

	fmt.Println("Deleted user with email", email)
}

func create(email, password string, conn *sql.Conn) {
	if len(password) == 0 {
		log.Fatal("Make sure to pass a password")
	}

	encrPW, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal(err)
	}

	query := `
		INSERT INTO users (email, password, created_at)
		VALUES ($1, $2, $3)
	`

	ctx := context.Background()
	res, err := conn.ExecContext(ctx, query, email, encrPW, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if nRows != 1 {
		log.Fatalf("expected 1 row to be inserted, Got: %v", nRows)
	}

	fmt.Println("Created user", email)
}
