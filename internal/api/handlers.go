package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"pkoforum/db"
	sqlcdb "pkoforum/db/sqlc"

	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai"
)

type Thread struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Comments  []Comment `json:"comments,omitempty"`
}

type Comment struct {
	ID        string            `json:"id"`
	ThreadID  string            `json:"thread_id"`
	Content   map[string]string `json:"content"`
	ImagePath string            `json:"image_path,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type LocalizedThread struct {
	ID        string             `json:"id"`
	Title     string             `json:"title"`
	Content   string             `json:"content"`
	CreatedAt time.Time          `json:"created_at"`
	Comments  []LocalizedComment `json:"comments,omitempty"`
	Language  string             `json:"language"`
}

type LocalizedComment struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	Content   string    `json:"content"`
	ImagePath string    `json:"image_path,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Language  string    `json:"language"`
}

var queries *sqlcdb.Queries
var openaiClient *openai.Client

func Init(q *sqlcdb.Queries) {
	queries = q
}

func InitOpenAI(client *openai.Client) {
	openaiClient = client
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

func GetThreads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	lang := GetLanguage(ctx)

	threads, err := queries.ListThreads(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var displayThreads []LocalizedThread
	for _, t := range threads {
		displayThreads = append(displayThreads, LocalizedThread{
			ID:        t.ID,
			Title:     t.Title,
			Content:   t.Content,
			CreatedAt: t.CreatedAt,
			Language:  lang,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(displayThreads)
}

func GetThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	lang := GetLanguage(ctx)
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

		fmt.Println(comment.Content)
		// Add translation if available
		if c.Language.Valid && c.Content.Valid {
			comment.Content[c.Language.String] = c.Content.String
		}

		// Add image if available
		if c.ImageID.Valid && c.Filepath.Valid {
			comment.ImagePath = c.Filepath.String
		}
	}

	displayThread := LocalizedThread{
		ID:        thread.ID,
		Title:     thread.Title,
		Content:   thread.Content,
		CreatedAt: thread.CreatedAt,
		Language:  lang,
	}

	// Convert comments to localized format
	for _, comment := range commentMap {
		localizedComment := LocalizedComment{
			ID:        comment.ID,
			ThreadID:  comment.ThreadID,
			Content:   GetLocalizedContent(comment.Content, lang),
			ImagePath: comment.ImagePath,
			CreatedAt: comment.CreatedAt,
			Language:  lang,
		}
		displayThread.Comments = append(displayThread.Comments, localizedComment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(displayThread)
}

func CreateThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	threadParams := sqlcdb.CreateThreadParams{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	thread, err := queries.CreateThread(ctx, threadParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	displayThread := Thread{
		ID:        thread.ID,
		Title:     thread.Title,
		Content:   thread.Content,
		CreatedAt: thread.CreatedAt,
		Comments:  []Comment{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(displayThread)
}

// processCommentTranslationInBackground handles the translation and saving of translations in the background
func processCommentTranslationInBackground(ctx context.Context, commentID string, originalContent string, isRussian bool) {
	// Create a new background context since the request context will be cancelled
	bgCtx := context.Background()

	// Start a new transaction for the background work
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for background translation: %v", err)
		return
	}
	defer tx.Rollback()

	// Create a new Queries instance with the transaction
	qtx := queries.WithTx(tx)

	// Only translate to the missing language
	var targetLang string
	if isRussian {
		targetLang = "en"
	} else {
		targetLang = "ru"
	}

	// Translate to the target language
	translation, err := translateText(bgCtx, originalContent, targetLang)
	if err != nil {
		log.Printf("Error translating to %s: %v", targetLang, err)
		return
	}

	// Save the translation
	_, err = qtx.CreateCommentTranslation(bgCtx, sqlcdb.CreateCommentTranslationParams{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		CommentID: commentID,
		Language:  targetLang,
		Content:   translation,
	})
	if err != nil {
		log.Printf("Error saving translation for language %s: %v", targetLang, err)
		return
	}

	log.Printf("Translation saved for comment %s in language %s", commentID, targetLang)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing translation transaction: %v", err)
		return
	}
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	threadID := vars["id"]

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	commentID := fmt.Sprintf("%d", time.Now().UnixNano())
	originalContent := r.FormValue("content")

	// Start a transaction for the initial comment creation
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

	// Detect if the text is Russian
	isRussian := false
	for _, r := range originalContent {
		if r >= 0x0400 && r <= 0x04FF {
			isRussian = true
			break
		}
	}

	// Save the original content translation immediately
	originalLang := "en"
	if isRussian {
		originalLang = "ru"
	}
	_, err = qtx.CreateCommentTranslation(ctx, sqlcdb.CreateCommentTranslationParams{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		CommentID: comment.ID,
		Language:  originalLang,
		Content:   originalContent,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize translations map with just the original content
	translations := make(map[string]string)
	translations[originalLang] = originalContent

	// Handle image upload if present
	var imagePath string
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
		imagePath = webPath
	}

	// Commit the initial transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Start background translation
	go processCommentTranslationInBackground(ctx, comment.ID, originalContent, isRussian)

	// Create response with just the original content
	response := Comment{
		ID:        comment.ID,
		ThreadID:  comment.ThreadID,
		Content:   translations,
		ImagePath: imagePath,
		CreatedAt: comment.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
