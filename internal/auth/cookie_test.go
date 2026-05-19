package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestManager_GenerateAndParseToken(t *testing.T) {
	cfg := Config{Secret: "testsecret", TokenTTL: time.Minute}
	m, _ := New(cfg)

	userID := "user123"
	t.Run("valid token", func(t *testing.T) {
		token, err := m.GenerateToken(userID)
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		gotUserID, err := m.ParseToken(token)
		if err != nil {
			t.Fatalf("ParseToken() error = %v", err)
		}

		if gotUserID != userID {
			t.Errorf("ParseToken() = %v, want %v", gotUserID, userID)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := m.ParseToken("invalid.token.string")
		if err != ErrInvalidToken {
			t.Errorf("ParseToken() error = %v, want ErrInvalidToken", err)
		}
	})
}

func TestManager_SetAndGetCookie(t *testing.T) {
	cfg := Config{Secret: "testsecret", TokenTTL: time.Minute, CookieName: "auth"}
	m, _ := New(cfg)

	userID := "user123"
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("set cookie", func(t *testing.T) {
		err := m.SetCookie(rec, userID)
		if err != nil {
			t.Fatalf("SetCookie() error = %v", err)
		}

		cookies := rec.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("SetCookie() did not set any cookie")
		}

		found := false
		for _, c := range cookies {
			if c.Name == m.cookieName && c.Value != "" {
				found = true
			}
		}

		if !found {
			t.Errorf("Cookie %v not set properly", m.cookieName)
		}
	})

	t.Run("get userID from cookie", func(t *testing.T) {
		for _, c := range rec.Result().Cookies() {
			req.AddCookie(c)
		}

		gotUserID, err := m.GetUserID(req)
		if err != nil {
			t.Fatalf("GetUserID() error = %v", err)
		}

		if gotUserID != userID {
			t.Errorf("GetUserID() = %v, want %v", gotUserID, userID)
		}
	})
}

func TestManager_GetOrCreateUserID(t *testing.T) {
	cfg := Config{Secret: "testsecret", TokenTTL: time.Minute, CookieName: "auth"}
	m, _ := New(cfg)

	t.Run("create new userID when no cookie", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		userID, err := m.GetOrCreateUserID(rec, req)
		if err != nil {
			t.Fatalf("GetOrCreateUserID() error = %v", err)
		}

		if userID == "" {
			t.Fatal("GetOrCreateUserID() returned empty userID")
		}

		cookies := rec.Result().Cookies()
		found := false
		for _, c := range cookies {
			if c.Name == m.cookieName && strings.Contains(c.Value, ".") {
				found = true
			}
		}
		if !found {
			t.Errorf("Cookie %v not set", m.cookieName)
		}
	})

	t.Run("return existing userID from cookie", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		userID1, _ := m.GetOrCreateUserID(rec, req)
		for _, c := range rec.Result().Cookies() {
			req.AddCookie(c)
		}

		userID2, err := m.GetOrCreateUserID(rec, req)
		if err != nil {
			t.Fatalf("GetOrCreateUserID() error = %v", err)
		}

		if userID1 != userID2 {
			t.Errorf("GetOrCreateUserID() = %v, want %v", userID2, userID1)
		}
	})
}
