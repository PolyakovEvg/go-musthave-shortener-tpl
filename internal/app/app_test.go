package app

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestNew_MemoryStorage(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":0",
		BaseURL:       "http://localhost:8080/",
		AuthSecret:    "testsecret",
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if app == nil {
		t.Fatal("expected app, got nil")
	}

	if app.server == nil {
		t.Error("expected server, got nil")
	}

	if app.logger == nil {
		t.Error("expected logger, got nil")
	}

	if app.Deleter == nil {
		t.Error("expected deleter, got nil")
	}

	app.Deleter.Close()
}

func TestNew_WithFilePath(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":0",
		BaseURL:       "http://localhost:8080/",
		AuthSecret:    "testsecret",
		FilePath:      t.TempDir() + "/test.json",
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if app == nil {
		t.Fatal("expected app, got nil")
	}

	app.Deleter.Close()
}

func TestApp_ServerHandler(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":0",
		BaseURL:       "http://localhost:8080/",
		AuthSecret:    "testsecret",
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer app.Deleter.Close()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	app.server.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestApp_ShortenURL(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":0",
		BaseURL:       "http://localhost:8080/",
		AuthSecret:    "testsecret",
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer app.Deleter.Close()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	app.server.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestApp_APIShorten(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":0",
		BaseURL:       "http://localhost:8080/",
		AuthSecret:    "testsecret",
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer app.Deleter.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/shortener", strings.NewReader(`{"url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.server.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestApp_MultipleRequests(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":0",
		BaseURL:       "http://localhost:8080/",
		AuthSecret:    "testsecret",
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer app.Deleter.Close()

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rr := httptest.NewRecorder()
		app.server.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i, http.StatusOK, rr.Code)
		}
	}
}

func TestApp_GzipMiddleware(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":0",
		BaseURL:       "http://localhost:8080/",
		AuthSecret:    "testsecret",
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer app.Deleter.Close()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	app.server.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestApp_Run_ServerStart(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":0",
		BaseURL:       "http://localhost:8080/",
		AuthSecret:    "testsecret",
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var g errgroup.Group

	g.Go(func() error {
		if err := app.Run(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	g.Go(func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		return app.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
