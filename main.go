package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"pkoforum/db"
	sqlcdb "pkoforum/db/sqlc"
	"pkoforum/internal/api"
	"pkoforum/internal/config"

	"github.com/gorilla/handlers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.CloseDB()

	// Initialize queries
	queries := sqlcdb.New(db.DB)

	// Initialize OpenAI client
	config := openai.DefaultConfig(cfg.DeepseekAPIKey)
	config.BaseURL = cfg.DeepseekURL
	openaiClient := openai.NewClientWithConfig(config)

	// Create uploads directory with proper permissions
	if err := os.MkdirAll(cfg.UploadsPath, 0755); err != nil {
		log.Fatal().Err(err).Str("path", cfg.UploadsPath).Msg("Failed to create uploads directory")
	}

	// Initialize the application
	app := api.NewApp(db.DB, queries, openaiClient, cfg.UploadsPath)
	router := app.Router()

	// CORS middleware
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Accept", "Accept-Language", "Accept-Encoding", "X-CSRF-Token", "Authorization"}),
		handlers.AllowCredentials(),
	)
	router.Use(corsMiddleware)
	router.Use(api.LanguageMiddleware)

	// Serve static files with proper headers
	fs := http.FileServer(http.Dir("static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) == ".jpg" || filepath.Ext(r.URL.Path) == ".jpeg" || filepath.Ext(r.URL.Path) == ".png" {
			w.Header().Set("Content-Type", "image/"+filepath.Ext(r.URL.Path)[1:])
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w, r)
	})))

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info().Str("addr", addr).Msg("Starting server")
	log.Fatal().Err(http.ListenAndServe(addr, router)).Msg("Server stopped")
}
