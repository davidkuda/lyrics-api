package config

import (
	"database/sql"
	"log"
)

type AppConfig struct {
	Logger *log.Logger
	DB     *sql.DB
	CORS   struct {
		TrustedOrigins []string
	}
}
