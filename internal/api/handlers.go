package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	sqlcdb "pkoforum/db/sqlc"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Thread struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
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
	Category  string             `json:"category"`
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

type CategoryOption struct {
	Value string            `json:"value"`
	Label map[string]string `json:"label"`
}

type LocalizedCategory struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// GetThreads handles the GET /api/threads endpoint
func (app *App) GetThreads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	lang := GetLanguage(ctx)

	category := r.URL.Query().Get("category")
	var threads []sqlcdb.Thread
	var err error

	log.Debug().Str("category", category).Msg("Getting threads")

	if category != "" {
		if !ValidateCategory(category) {
			http.Error(w, "Invalid category", http.StatusBadRequest)
			return
		}

		threads, err = app.queries.ListThreads(ctx, category)
		if err != nil {
			log.Error().Err(err).Str("category", category).Msg("Error listing threads")
			threads = []sqlcdb.Thread{}
		}
	} else {
		threads, err = app.queries.ListAllThreads(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Error listing all threads")
			threads = []sqlcdb.Thread{}
		}
	}

	log.Debug().Int("count", len(threads)).Str("category", category).Msg("Found threads")

	displayThreads := make([]LocalizedThread, 0)

	for _, t := range threads {
		displayThreads = append(displayThreads, LocalizedThread{
			ID:        t.ID,
			Title:     t.Title,
			Content:   t.Content,
			Category:  t.Category,
			CreatedAt: t.CreatedAt,
			Language:  lang,
			Comments:  make([]LocalizedComment, 0),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(displayThreads)
}

// GetThread handles the GET /api/threads/{id} endpoint
func (app *App) GetThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	lang := GetLanguage(ctx)
	vars := mux.Vars(r)
	threadID := vars["id"]

	thread, err := app.queries.GetThread(ctx, threadID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Str("thread_id", threadID).Msg("Thread not found")
			http.Error(w, "Thread not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Str("thread_id", threadID).Msg("Error getting thread")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	comments, err := app.queries.GetThreadComments(ctx, threadID)
	if err != nil {
		log.Error().Err(err).Str("thread_id", threadID).Msg("Error getting thread comments")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

		if c.Language.Valid && c.Content.Valid {
			comment.Content[c.Language.String] = c.Content.String
		}

		if c.ImageID.Valid && c.Filepath.Valid {
			comment.ImagePath = c.Filepath.String
		}
	}

	displayThread := LocalizedThread{
		ID:        thread.ID,
		Title:     thread.Title,
		Content:   thread.Content,
		Category:  thread.Category,
		CreatedAt: thread.CreatedAt,
		Language:  lang,
	}

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

// CreateThread handles the POST /api/threads endpoint
func (app *App) CreateThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Error decoding request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Category == "" {
		log.Debug().Msg("Category is required")
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}

	if !ValidateCategory(req.Category) {
		log.Debug().Str("category", req.Category).Msg("Invalid category")
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	}

	threadParams := sqlcdb.CreateThreadParams{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Title:     req.Title,
		Content:   req.Content,
		Category:  req.Category,
		CreatedAt: time.Now(),
	}

	thread, err := app.queries.CreateThread(ctx, threadParams)
	if err != nil {
		log.Error().Err(err).Interface("params", threadParams).Msg("Error creating thread")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Str("thread_id", thread.ID).Str("category", thread.Category).Msg("Thread created")

	displayThread := Thread{
		ID:        thread.ID,
		Title:     thread.Title,
		Content:   thread.Content,
		Category:  thread.Category,
		CreatedAt: thread.CreatedAt,
		Comments:  []Comment{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(displayThread)
}

// processCommentTranslationInBackground handles the translation and saving of translations
func (app *App) processCommentTranslationInBackground(ctx context.Context, commentID string, originalContent string, isRussian bool) {
	bgCtx := context.Background()

	tx, err := app.db.Begin()
	if err != nil {
		log.Error().Err(err).Str("comment_id", commentID).Msg("Error starting transaction for translation")
		return
	}
	defer tx.Rollback()

	qtx := app.queries.WithTx(tx)

	var targetLang string
	if isRussian {
		targetLang = "en"
	} else {
		targetLang = "ru"
	}

	translation, err := app.translateText(bgCtx, originalContent, targetLang)
	if err != nil {
		log.Error().Err(err).
			Str("comment_id", commentID).
			Str("target_lang", targetLang).
			Msg("Error translating content")
		return
	}

	_, err = qtx.CreateCommentTranslation(bgCtx, sqlcdb.CreateCommentTranslationParams{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		CommentID: commentID,
		Language:  targetLang,
		Content:   translation,
	})
	if err != nil {
		log.Error().Err(err).
			Str("comment_id", commentID).
			Str("target_lang", targetLang).
			Msg("Error saving translation")
		return
	}

	log.Info().
		Str("comment_id", commentID).
		Str("target_lang", targetLang).
		Msg("Translation saved")

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).
			Str("comment_id", commentID).
			Msg("Error committing translation transaction")
		return
	}
}

// CreateComment handles the POST /api/threads/{id}/comments endpoint
func (app *App) CreateComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	threadID := vars["id"]

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Error().Err(err).Msg("Error parsing multipart form")
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	commentID := fmt.Sprintf("%d", time.Now().UnixNano())
	originalContent := r.FormValue("content")

	tx, err := app.db.Begin()
	if err != nil {
		log.Error().Err(err).Msg("Error starting transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := app.queries.WithTx(tx)

	comment, err := qtx.CreateComment(ctx, sqlcdb.CreateCommentParams{
		ID:        commentID,
		ThreadID:  threadID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Error().Err(err).
			Str("thread_id", threadID).
			Str("comment_id", commentID).
			Msg("Error creating comment")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isRussian := false
	for _, r := range originalContent {
		if r >= 0x0400 && r <= 0x04FF {
			isRussian = true
			break
		}
	}

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
		log.Error().Err(err).
			Str("comment_id", comment.ID).
			Str("language", originalLang).
			Msg("Error creating comment translation")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	translations := make(map[string]string)
	translations[originalLang] = originalContent

	var imagePath string
	file, header, err := r.FormFile("image")
	if err == nil && file != nil {
		defer file.Close()

		if err := os.MkdirAll(app.uploadsPath, 0755); err != nil {
			log.Error().Err(err).Str("path", app.uploadsPath).Msg("Error creating uploads directory")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
		filepath := filepath.Join(app.uploadsPath, filename)

		dst, err := os.Create(filepath)
		if err != nil {
			log.Error().Err(err).Str("path", filepath).Msg("Error creating file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			log.Error().Err(err).Str("path", filepath).Msg("Error copying file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		webPath := "/static/uploads/" + filename
		_, err = qtx.CreateCommentImage(ctx, sqlcdb.CreateCommentImageParams{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			CommentID: comment.ID,
			Filename:  filename,
			Filepath:  webPath,
			CreatedAt: time.Now(),
		})
		if err != nil {
			log.Error().Err(err).
				Str("comment_id", comment.ID).
				Str("filename", filename).
				Msg("Error creating comment image")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		imagePath = webPath
		log.Debug().
			Str("comment_id", comment.ID).
			Str("path", webPath).
			Msg("Image uploaded")
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Str("comment_id", comment.ID).Msg("Error committing transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go app.processCommentTranslationInBackground(ctx, comment.ID, originalContent, isRussian)

	log.Info().
		Str("comment_id", comment.ID).
		Str("thread_id", comment.ThreadID).
		Bool("has_image", imagePath != "").
		Msg("Comment created")

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

// GetCategories handles the GET /api/categories endpoint
func (app *App) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	lang := GetLanguage(ctx)

	categories := GetLocalizedCategories()

	var localizedCategories []LocalizedCategory
	for _, cat := range categories {
		label := cat.Label[lang]
		if label == "" {
			label = cat.Label["en"]
		}

		localizedCategories = append(localizedCategories, LocalizedCategory{
			Value: cat.Value,
			Label: label,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(localizedCategories)
}

// GetLocalizedCategories returns the list of categories with translations
func GetLocalizedCategories() []CategoryOption {
	return []CategoryOption{
		{
			Value: "general",
			Label: map[string]string{
				"en": "General",
				"ru": "Общее",
			},
		},
		{
			Value: "help",
			Label: map[string]string{
				"en": "Help",
				"ru": "Помощь",
			},
		},
		{
			Value: "discussion",
			Label: map[string]string{
				"en": "Discussion",
				"ru": "Обсуждение",
			},
		},
		{
			Value: "announcement",
			Label: map[string]string{
				"en": "Announcement",
				"ru": "Объявление",
			},
		},
	}
}

// ValidateCategory checks if a category is valid
func ValidateCategory(category string) bool {
	validCategories := GetLocalizedCategories()
	for _, cat := range validCategories {
		if cat.Value == category {
			return true
		}
	}
	return false
}
