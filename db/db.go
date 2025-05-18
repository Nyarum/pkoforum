package db

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() error {
	// Ensure the data directory exists
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Error().Err(err).Str("path", dataDir).Msg("Failed to create data directory")
		return err
	}

	// Open SQLite database
	dbPath := filepath.Join(dataDir, "forum.db")
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Error().Err(err).Str("path", dbPath).Msg("Failed to open database")
		return err
	}

	// Create tables
	schema, err := os.ReadFile("db/sqlc/schema.sql")
	if err != nil {
		log.Error().Err(err).Msg("Failed to read schema file")
		return err
	}

	_, err = DB.Exec(string(schema))
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute schema")
		return err
	}

	log.Info().Str("path", dbPath).Msg("Database initialized successfully")
	return nil
}

func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing database connection")
		} else {
			log.Info().Msg("Database connection closed successfully")
		}
	}
}
