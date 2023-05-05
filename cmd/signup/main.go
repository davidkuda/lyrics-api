package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/davidkuda/lyricsapi/dbio"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	conn := DBConn()
	defer conn.Close()

	userName := flag.String("user-name", "", "The name of the new user")
	password := flag.String("password", "", "The password of the new user")
	deleteUser := flag.String("delete-user", "", "A user that should be removed from the DB")
	flag.Parse()

	if *userName != "" && *password != "" {
		create(*userName, *password, conn)
		return
	}

	if *deleteUser != "" {
		delete(*userName, conn)
		return
	}

	log.Fatal("Did you use the CLI correctly? Nothing happened.")
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
			WHERE name = $1
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

func create(userName, password string, conn *sql.Conn) {
	if len(password) == 0 {
		log.Fatal("Make sure to pass a password")
	}

	encrPW, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal(err)
	}

	query := `
		INSERT INTO users (name, password)
		VALUES ($1, $2)
	`

	ctx := context.Background()
	res, err := conn.ExecContext(ctx, query, userName, encrPW)
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

	fmt.Println("Created user", userName)
}
