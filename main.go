package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"pkoforum/db"
	sqlcdb "pkoforum/db/sqlc"
	"pkoforum/internal/api"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai"
)

var queries *sqlcdb.Queries

func init() {
	// Initialize database
	db.InitDB()

	// Initialize queries
	queries = sqlcdb.New(db.DB)

	// Initialize OpenAI client
	config := openai.DefaultConfig("sk-7fecdc2078a344b0ae899c243fe8b5fb")
	config.BaseURL = "https://api.deepseek.com"
	openaiClient := openai.NewClientWithConfig(config)

	// Initialize API handlers
	api.Init(queries)
	api.InitOpenAI(openaiClient)

	// Create uploads directory with proper permissions
	uploadsDir := "static/uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer db.CloseDB()

	r := mux.NewRouter()

	// CORS middleware
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Accept", "Accept-Language", "Accept-Encoding", "X-CSRF-Token", "Authorization"}),
		handlers.AllowCredentials(),
	)
	r.Use(corsMiddleware)
	r.Use(api.LanguageMiddleware)

	// Serve static files with proper headers
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set proper headers for images
		if filepath.Ext(r.URL.Path) == ".jpg" || filepath.Ext(r.URL.Path) == ".jpeg" || filepath.Ext(r.URL.Path) == ".png" {
			w.Header().Set("Content-Type", "image/"+filepath.Ext(r.URL.Path)[1:])
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w, r)
	})))

	// API Routes
	r.HandleFunc("/api/threads", api.GetThreads).Methods("GET")
	r.HandleFunc("/api/threads", api.CreateThread).Methods("POST")
	r.HandleFunc("/api/threads/{id}", api.GetThread).Methods("GET")
	r.HandleFunc("/api/threads/{id}/comments", api.CreateComment).Methods("POST")

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
