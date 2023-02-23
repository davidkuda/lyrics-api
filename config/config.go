package config

import (
	"database/sql"
	"log"
)

type AppConfig struct {
	logger *log.Logger
	db     *sql.DB
}
