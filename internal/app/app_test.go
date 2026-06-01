package app

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		t.Errorf("shutdown error: %v", err)
	}

	app.Deleter.Close()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("server did not shut down in time")
	}
}
