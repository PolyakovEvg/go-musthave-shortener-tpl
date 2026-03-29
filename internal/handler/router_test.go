package handler

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/auth"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	authmw "PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/auth"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/memory"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/service/url"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func newTestRouter(cfg *config.Config) *chi.Mux {
	repo := memory.New()
	svc := url.NewURLService(repo, cfg.BaseURL)

	h := NewURLHandler(svc, cfg, nil)

	r := chi.NewRouter()

	authMgr, _ := auth.New(auth.Config{Secret: "testsecret", TokenTTL: 24 * time.Hour})
	r.Use(authmw.WithCookie(authMgr))

	h.Register(r)
	return r
}

func TestRouter(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080/",
	}

	mux := newTestRouter(cfg)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		contentType    string
		expectedStatus int
		checkBody      bool
	}{
		{
			name:           "POST valid URL",
			method:         http.MethodPost,
			path:           "/",
			body:           "https://example.com",
			contentType:    "text/plain",
			expectedStatus: http.StatusCreated,
			checkBody:      true,
		},
		{
			name:           "POST wrong content type",
			method:         http.MethodPost,
			path:           "/",
			body:           "https://example.com",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkBody:      false,
		},
		{
			name:           "GET root path should return bad request",
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusBadRequest,
			checkBody:      false,
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			path:           "/",
			expectedStatus: http.StatusBadRequest,
			checkBody:      false,
		},
		{
			name:           "DELETE method not allowed",
			method:         http.MethodDelete,
			path:           "/",
			expectedStatus: http.StatusBadRequest,
			checkBody:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkBody && tt.expectedStatus == http.StatusCreated {
				body := rr.Body.String()

				if !strings.HasPrefix(body, cfg.BaseURL) {
					t.Errorf("expected body to start with %s, got %s", cfg.BaseURL, body)
				}
			}
		})
	}
}

func TestRouter_Integration_ShortenAndRedirect(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080/",
	}

	mux := newTestRouter(cfg)

	originalURL := "https://github.com/PolyakovEvg/go-musthave-shortener-tpl"
	req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	req1.Header.Set("Content-Type", "text/plain")

	rr1 := httptest.NewRecorder()
	mux.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusCreated {
		t.Fatalf("POST expected status %d, got %d", http.StatusCreated, rr1.Code)
	}

	shortURL := rr1.Body.String()
	if !strings.HasPrefix(shortURL, cfg.BaseURL) {
		t.Fatalf("short URL should start with base URL, got: %s", shortURL)
	}

	shortID := strings.TrimPrefix(shortURL, cfg.BaseURL)

	req2 := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTemporaryRedirect {
		t.Errorf("GET expected status %d, got %d", http.StatusTemporaryRedirect, rr2.Code)
	}

	location := rr2.Header().Get("Location")
	if location != originalURL {
		t.Errorf("expected Location header %q, got %q", originalURL, location)
	}
}

func TestRouter_MultipleURLs(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080/",
	}

	mux := newTestRouter(cfg)

	urls := []string{
		"https://example.com//page1",
		"https://example.com/page2",
		"https://example.com/page3",
	}

	shortIDs := make([]string, len(urls))

	for i, url := range urls {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url))
		req.Header.Set("Content-Type", "text/plain")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("POST failed for URL %d: status %d", i, rr.Code)
		}

		shortURL := rr.Body.String()
		shortID := strings.TrimPrefix(shortURL, cfg.BaseURL)
		shortIDs[i] = shortID
	}

	seen := make(map[string]bool)
	for i, shortID := range shortIDs {
		if seen[shortID] {
			t.Errorf("duplicate short ID generated: %s", shortID)
		}
		seen[shortID] = true

		req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusTemporaryRedirect {
			t.Errorf("GET failed for short ID %s: status %d", shortID, rr.Code)
		}

		location := rr.Header().Get("Location")
		if location != urls[i] {
			t.Errorf("expected %q, got %q", urls[i], location)
		}
	}
}

func TestAPI_ShortenURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080/",
	}
	mux := newTestRouter(cfg)

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		contentType    string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:   "Valid JSON request",
			method: http.MethodPost,
			path:   "/api/shortener",
			body: shortenRequest{
				URL: "https://example.com",
			},
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name:           "Wrong content type",
			method:         http.MethodPost,
			path:           "/api/shortener",
			body:           "https://example.com",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			path:           "/api/shortener",
			body:           `{"invalid": "json"`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:   "Empty URL",
			method: http.MethodPost,
			path:   "/api/shortener",
			body: shortenRequest{
				URL: "",
			},
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:           "Missing URL field",
			method:         http.MethodPost,
			path:           "/api/shortener",
			body:           map[string]string{},
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:           "Empty body",
			method:         http.MethodPost,
			path:           "/api/shortener",
			body:           "",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:   "URL with spaces",
			method: http.MethodPost,
			path:   "/api/shortener",
			body: shortenRequest{
				URL: "  https://example.com  ",
			},
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody string
			switch v := tt.body.(type) {
			case string:
				reqBody = v
			default:
				jsonBody, _ := json.Marshal(v)
				reqBody = string(jsonBody)
			}

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(reqBody))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse && tt.expectedStatus == http.StatusCreated {
				var resp shortenResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				if err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if !strings.HasPrefix(resp.Result, cfg.BaseURL) {
					t.Errorf("expected result to start with %s, got %s", cfg.BaseURL, resp.Result)
				}

				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", contentType)
				}
			}
		})
	}
}
