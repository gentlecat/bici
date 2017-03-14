package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	// These are database connection details that are meant to be used
	// with the Docker setup.
	DB_DRIVER   = "postgres"
	DB_HOST     = "db"
	DB_PORT     = 5432
	DB_USER     = "summits"
	DB_PASSWORD = "summits"
	DB_NAME     = "summits"
)

var (
	db *sql.DB
)

func EstablishConnection() (err error) {
	db, err = sql.Open(DB_DRIVER, fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME))
	return err
}
