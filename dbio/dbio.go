package dbio

import (
	"database/sql"
	"net/url"

	_ "github.com/jackc/pgx/v5/stdlib"
)

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
