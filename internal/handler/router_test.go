package handler

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouter(t *testing.T) {
	storage := repository.NewStorage()
	cfg := &config.Config{
		BaseURL: "http://localhost:8080/",
	}

	mux := Router(storage, cfg)

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
			expectedStatus: http.StatusUnsupportedMediaType,
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
	storage := repository.NewStorage()
	cfg := &config.Config{
		BaseURL: "http://test-server:8080/",
	}

	mux := Router(storage, cfg)

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
	storage := repository.NewStorage()
	cfg := &config.Config{
		BaseURL: "http://localhost/",
	}

	mux := Router(storage, cfg)

	urls := []string{
		"https://example.com/page1",
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
