package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"pkoforum/db"
	sqlcdb "pkoforum/db/sqlc"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai"
)

type Comment struct {
	ID        string            `json:"id"`
	ThreadID  string            `json:"threadId"`
	Content   map[string]string `json:"content"` // Map of language code to content
	ImagePath string            `json:"imagePath,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
}

type Thread struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	Comments  []Comment `json:"comments"`
}

type ThreadStore struct {
	sync.RWMutex
	threads []Thread
}

type PageData struct {
	Threads          []Thread
	SelectedThread   *Thread
	SelectedThreadID string
}

var store = ThreadStore{
	threads: make([]Thread, 0),
}

var templates *template.Template
var queries *sqlcdb.Queries
var openaiClient *openai.Client

func init() {
	// Create template functions map
	funcMap := template.FuncMap{
		"json": func(v interface{}) template.JS {
			a, err := json.Marshal(v)
			if err != nil {
				log.Printf("Error marshaling to JSON: %v", err)
				return template.JS("{}")
			}
			return template.JS(a)
		},
	}

	// Initialize templates with function map
	templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	// Initialize database
	db.InitDB()
	queries = sqlcdb.New(db.DB)

	config := openai.DefaultConfig("sk-7fecdc2078a344b0ae899c243fe8b5fb")
	config.BaseURL = "https://api.deepseek.com"

	// Initialize OpenAI client
	openaiClient = openai.NewClientWithConfig(config)

	// Create uploads directory
	if err := os.MkdirAll("static/uploads", 0755); err != nil {
		log.Fatal(err)
	}
}

// translateText translates text between English and Russian using OpenAI
func translateText(ctx context.Context, text, targetLang string) (string, error) {
	var prompt string
	if targetLang == "ru" {
		prompt = fmt.Sprintf("Translate the following English text to Russian:\n\n%s\n\nAnswer with only translated variant without anything else, if you can't translate, return the original text", text)
	} else {
		prompt = fmt.Sprintf("Translate the following Russian text to English:\n\n%s\n\nAnswer with only translated variant without anything else, if you can't translate, return the original text", text)
	}

	resp, err := openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "deepseek-chat",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("translation error: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no translation received")
	}

	return resp.Choices[0].Message.Content, nil
}

func main() {
	defer db.CloseDB()

	r := mux.NewRouter()

	// CORS middleware
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Accept", "Accept-Encoding", "X-CSRF-Token", "Authorization"}),
		handlers.AllowCredentials(),
	)
	r.Use(corsMiddleware)

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	r.HandleFunc("/", handleHome).Methods("GET")
	r.HandleFunc("/api/threads", handleCreateThread).Methods("POST")
	r.HandleFunc("/api/threads", handleGetThreads).Methods("GET")
	r.HandleFunc("/api/threads/{id}", handleGetThread).Methods("GET")
	r.HandleFunc("/api/threads/{id}/comments", handleAddComment).Methods("POST")

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threads, err := queries.ListThreads(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var displayThreads []Thread
	for _, t := range threads {
		displayThreads = append(displayThreads, Thread{
			ID:        t.ID,
			Title:     t.Title,
			Content:   t.Content,
			CreatedAt: t.CreatedAt,
		})
	}

	data := PageData{
		Threads: displayThreads,
	}

	if err := templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleCreateThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
	}

	threadParams := sqlcdb.CreateThreadParams{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Title:     r.FormValue("title"),
		Content:   r.FormValue("content"),
		CreatedAt: time.Now(),
	}

	_, err := queries.CreateThread(ctx, threadParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	threads, err := queries.ListThreads(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var displayThreads []Thread
	for _, t := range threads {
		displayThreads = append(displayThreads, Thread{
			ID:        t.ID,
			Title:     t.Title,
			Content:   t.Content,
			CreatedAt: t.CreatedAt,
		})
	}

	if err := templates.ExecuteTemplate(w, "threads", PageData{Threads: displayThreads}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGetThreads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threads, err := queries.ListThreads(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var displayThreads []Thread
	for _, t := range threads {
		displayThreads = append(displayThreads, Thread{
			ID:        t.ID,
			Title:     t.Title,
			Content:   t.Content,
			CreatedAt: t.CreatedAt,
		})
	}

	if err := templates.ExecuteTemplate(w, "threads", PageData{Threads: displayThreads}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGetThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	threadID := vars["id"]

	thread, err := queries.GetThread(ctx, threadID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Thread not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	comments, err := queries.GetThreadComments(ctx, threadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	displayThread := Thread{
		ID:        thread.ID,
		Title:     thread.Title,
		Content:   thread.Content,
		CreatedAt: thread.CreatedAt,
	}

	// Group comments by ID to combine translations
	commentMap := make(map[string]*Comment)
	for _, c := range comments {
		comment, exists := commentMap[c.ID]
		if !exists {
			comment = &Comment{
				ID:        c.ID,
				ThreadID:  c.ThreadID,
				Content:   make(map[string]string),
				CreatedAt: c.CreatedAt,
			}
			commentMap[c.ID] = comment
		}

		// Add translation if available
		if c.Language.Valid && c.Content.Valid {
			comment.Content[c.Language.String] = c.Content.String
		}

		// Add image if available
		if c.ImageID.Valid && c.Filepath.Valid {
			comment.ImagePath = c.Filepath.String
		}
	}

	// Convert map to slice
	for _, comment := range commentMap {
		displayThread.Comments = append(displayThread.Comments, *comment)
	}

	if err := templates.ExecuteTemplate(w, "thread", displayThread); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleAddComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	threadID := vars["id"]

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	commentID := fmt.Sprintf("%d", time.Now().UnixNano())
	originalContent := r.FormValue("content")

	// Start a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create a new Queries instance with the transaction
	qtx := queries.WithTx(tx)

	// Create the comment
	comment, err := qtx.CreateComment(ctx, sqlcdb.CreateCommentParams{
		ID:        commentID,
		ThreadID:  threadID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Detect the original language and translate to the other language
	isRussian := false
	for _, r := range originalContent {
		if r >= 0x0400 && r <= 0x04FF {
			isRussian = true
			break
		}
	}

	var translations = make(map[string]string)
	var translationErr error

	if isRussian {
		// Content is in Russian, translate to English
		translations["ru"] = originalContent
		translations["en"], translationErr = translateText(ctx, originalContent, "en")
	} else {
		// Content is in English (default), translate to Russian
		translations["en"] = originalContent
		translations["ru"], translationErr = translateText(ctx, originalContent, "ru")
	}

	if translationErr != nil {
		log.Printf("Translation error: %v", translationErr)
		// Continue with original content only
		translations["en"] = originalContent
		translations["ru"] = originalContent
	}

	// Save translations
	for lang, content := range translations {
		_, err = qtx.CreateCommentTranslation(ctx, sqlcdb.CreateCommentTranslationParams{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			CommentID: comment.ID,
			Language:  lang,
			Content:   content,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Handle image upload if present
	file, header, err := r.FormFile("image")
	if err == nil && file != nil {
		defer file.Close()

		// Create uploads directory if it doesn't exist
		uploadsDir := "static/uploads"
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Generate unique filename
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
		filepath := filepath.Join(uploadsDir, filename)

		// Create the file
		dst, err := os.Create(filepath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Save image info to database
		webPath := "/static/uploads/" + filename
		_, err = qtx.CreateCommentImage(ctx, sqlcdb.CreateCommentImageParams{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			CommentID: comment.ID,
			Filename:  filename,
			Filepath:  webPath,
			CreatedAt: time.Now(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated thread data
	thread, err := queries.GetThread(ctx, threadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all comments for the thread
	comments, err := queries.GetThreadComments(ctx, threadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	displayThread := Thread{
		ID:        thread.ID,
		Title:     thread.Title,
		Content:   thread.Content,
		CreatedAt: thread.CreatedAt,
	}

	// Group comments by ID to combine translations
	commentMap := make(map[string]*Comment)
	for _, c := range comments {
		comment, exists := commentMap[c.ID]
		if !exists {
			comment = &Comment{
				ID:        c.ID,
				ThreadID:  c.ThreadID,
				Content:   make(map[string]string),
				CreatedAt: c.CreatedAt,
			}
			commentMap[c.ID] = comment
		}

		// Add translation if available
		if c.Language.Valid && c.Content.Valid {
			comment.Content[c.Language.String] = c.Content.String
		}

		// Add image if available
		if c.ImageID.Valid && c.Filepath.Valid {
			comment.ImagePath = c.Filepath.String
		}
	}

	// Convert map to slice
	for _, comment := range commentMap {
		displayThread.Comments = append(displayThread.Comments, *comment)
	}

	if err := templates.ExecuteTemplate(w, "thread", displayThread); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
