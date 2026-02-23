package handler

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func (h *URLHandler) shortenURL(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "text/plain" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

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

func (h *URLHandler) redirectURL(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")

	shortID = strings.TrimPrefix(shortID, "/")

	fmt.Println("shortURL", shortID)
	originalURL, exists := h.storage.Get(shortID)
	if !exists {
		http.Error(w, "Not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *URLHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.shortenURL)
	r.Get("/{id}", h.redirectURL)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
	})
}

func Router(s *repository.Storage, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	handler := NewURLHandler(s, cfg.BaseURL)
	handler.RegisterRoutes(r)

	return r
}
