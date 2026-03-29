package handler

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	authmw "PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/auth"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/logger"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/service/url"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
)

type URLHandler struct {
	service *url.URLService
	config  *config.Config
	logger  *logger.Logger
}
type shortenRequest struct {
	URL string `json:"url"`
}
type shortenResponse struct {
	Result string `json:"result"`
}

func (h *URLHandler) Register(r chi.Router) {
	r.Post("/", h.shortenURL)
	r.Get("/{id}", h.redirectURL)
	r.Post("/{api}/{shorten}", h.postShorten)
	r.Get("/ping", h.PingHandler)
	r.Post("/{api}/{shorten}/{batch}", h.ShortenBatch)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	})
}

func NewURLHandler(svc *url.URLService, cfg *config.Config, logg *logger.Logger) *URLHandler {
	return &URLHandler{
		service: svc,
		config:  cfg,
		logger:  logg,
	}
}

func (h *URLHandler) shortenURL(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "text/plain" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, ok := authmw.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	shortURL, err := h.service.SaveShorten(userID, originalURL)

	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortURL))
			return
		}

		if errors.Is(err, os.ErrPermission) {
			h.logger.Zap.Errorw("permission denied",
				"file", h.config.FilePath,
				"error", err,
			)
		}

		h.logger.Zap.Errorw("failed to save shortened url",
			"url", originalURL,
			"error", err,
		)

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *URLHandler) redirectURL(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")

	shortID = strings.TrimPrefix(shortID, "/")
	originalURL, err := h.service.GetOriginal(shortID)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *URLHandler) postShorten(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")

	if ct != "application/json" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req shortenRequest
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, ok := authmw.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	shortURL, err := h.service.SaveShorten(userID, req.URL)

	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			resp := shortenResponse{
				Result: shortURL,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)

			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := shortenResponse{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
