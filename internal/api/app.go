package api

import (
	"context"
	"database/sql"
	sqlcdb "pkoforum/db/sqlc"

	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai"
)

// Querier defines the database operations interface
type Querier interface {
	CreateComment(ctx context.Context, arg sqlcdb.CreateCommentParams) (sqlcdb.Comment, error)
	CreateCommentImage(ctx context.Context, arg sqlcdb.CreateCommentImageParams) (sqlcdb.CommentImage, error)
	CreateCommentTranslation(ctx context.Context, arg sqlcdb.CreateCommentTranslationParams) (sqlcdb.CommentTranslation, error)
	CreateThread(ctx context.Context, arg sqlcdb.CreateThreadParams) (sqlcdb.Thread, error)
	GetThread(ctx context.Context, id string) (sqlcdb.Thread, error)
	GetThreadComments(ctx context.Context, threadID string) ([]sqlcdb.GetThreadCommentsRow, error)
	ListAllThreads(ctx context.Context) ([]sqlcdb.Thread, error)
	ListThreads(ctx context.Context, category string) ([]sqlcdb.Thread, error)
	WithTx(tx *sql.Tx) *sqlcdb.Queries
}

// App represents the application and its dependencies
type App struct {
	db          *sql.DB
	queries     Querier
	openai      *openai.Client
	router      *mux.Router
	uploadsPath string
}

// NewApp creates a new application instance
func NewApp(db *sql.DB, queries Querier, openai *openai.Client, uploadsPath string) *App {
	app := &App{
		db:          db,
		queries:     queries,
		openai:      openai,
		router:      mux.NewRouter(),
		uploadsPath: uploadsPath,
	}
	app.setupRoutes()
	return app
}

// setupRoutes configures all the routes for the application
func (app *App) setupRoutes() {
	// API Routes
	app.router.HandleFunc("/api/threads", app.GetThreads).Methods("GET")
	app.router.HandleFunc("/api/threads", app.CreateThread).Methods("POST")
	app.router.HandleFunc("/api/threads/{id}", app.GetThread).Methods("GET")
	app.router.HandleFunc("/api/threads/{id}/comments", app.CreateComment).Methods("POST")
	app.router.HandleFunc("/api/categories", app.GetCategories).Methods("GET")
}

// Router returns the configured router
func (app *App) Router() *mux.Router {
	return app.router
}
