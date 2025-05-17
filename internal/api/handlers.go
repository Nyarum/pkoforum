package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

type Thread struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Comments  []Comment `json:"comments,omitempty"`
}

type Comment struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	Content   string    `json:"content"`
	ImagePath string    `json:"image_path,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateThreadRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

func GetThreads(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement database query
	threads := []Thread{
		{
			ID:        "1",
			Title:     "Example Thread",
			Content:   "This is an example thread content",
			CreatedAt: time.Now(),
			Comments:  []Comment{},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(threads)
}

func GetThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID := vars["id"]

	// TODO: Implement database query
	thread := Thread{
		ID:        threadID,
		Title:     "Example Thread",
		Content:   "This is an example thread content",
		CreatedAt: time.Now(),
		Comments:  []Comment{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thread)
}

func CreateThread(w http.ResponseWriter, r *http.Request) {
	var req CreateThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Implement database insertion
	thread := Thread{
		ID:        "new-id", // Generate proper ID
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: time.Now(),
		Comments:  []Comment{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(thread)
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID := vars["id"]

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	file, handler, err := r.FormFile("image")
	var imagePath string

	if err == nil && file != nil {
		defer file.Close()

		// TODO: Implement proper file storage
		filename := filepath.Join("static", "uploads", handler.Filename)
		imagePath = "/" + filename
		// Save file logic here
	}

	// TODO: Implement database insertion
	comment := Comment{
		ID:        "new-id", // Generate proper ID
		ThreadID:  threadID,
		Content:   content,
		ImagePath: imagePath,
		CreatedAt: time.Now(),
	}

	// TODO: Fetch updated thread from database
	thread := Thread{
		ID:        threadID,
		Title:     "Example Thread",
		Content:   "This is an example thread content",
		CreatedAt: time.Now(),
		Comments:  []Comment{comment},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(thread)
}
