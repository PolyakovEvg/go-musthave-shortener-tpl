package handler

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"io"
	"net/http"
	"strings"
)

type URLHandler struct {
	storage *repository.Storage
	baseURL string
}

func NewURLHandler(storage *repository.Storage, baseURL string) *URLHandler {
	return &URLHandler{
		storage: storage,
		baseURL: baseURL,
	}
}

func (h *URLHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	originalURL = strings.TrimSpace(originalURL)

	shortID, err := h.storage.Save(originalURL)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	shortURL := h.baseURL + shortID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *URLHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	if path == "/" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	shortID := strings.TrimPrefix(path, "/")
	originalURL, exists := h.storage.Get(shortID)
	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URLHandler) MainHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodGet:
		h.handleGet(w, r)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func (h *URLHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.MainHandler)
}

func Router(s *repository.Storage, cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()
	handler := NewURLHandler(s, cfg.BaseURL)
	handler.RegisterRoutes(mux)
	return mux
}
