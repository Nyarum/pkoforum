package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	// Ensure the data directory exists
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Open SQLite database
	dbPath := filepath.Join(dataDir, "forum.db")
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create tables
	schema, err := os.ReadFile("db/sqlc/schema.sql")
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(string(schema))
	if err != nil {
		log.Fatal(err)
	}
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
