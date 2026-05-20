package auth

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/auth"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserIDFromContext(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		expectedID string
		expectedOK bool
	}{
		{
			name:       "context with user ID",
			ctx:        context.WithValue(context.Background(), userIDKey, "user123"),
			expectedID: "user123",
			expectedOK: true,
		},
		{
			name:       "context without user ID",
			ctx:        context.Background(),
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "context with wrong type",
			ctx:        context.WithValue(context.Background(), userIDKey, 123),
			expectedID: "",
			expectedOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := UserIDFromContext(tt.ctx)
			if id != tt.expectedID {
				t.Errorf("expected ID %q, got %q", tt.expectedID, id)
			}
			if ok != tt.expectedOK {
				t.Errorf("expected OK %v, got %v", tt.expectedOK, ok)
			}
		})
	}
}

func TestWithCookie(t *testing.T) {
	authMgr, err := auth.New(auth.Config{Secret: "testsecret", TokenTTL: 24 * time.Hour})
	if err != nil {
		t.Fatalf("failed to create auth manager: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected user ID in context")
			http.Error(w, "no user ID", http.StatusInternalServerError)
			return
		}
		if userID == "" {
			t.Error("expected non-empty user ID")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userID))
	})

	middleware := WithCookie(authMgr)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("expected cookie to be set")
	}
}

func TestWithCookie_ExistingUser(t *testing.T) {
	authMgr, err := auth.New(auth.Config{Secret: "testsecret", TokenTTL: 24 * time.Hour})
	if err != nil {
		t.Fatalf("failed to create auth manager: %v", err)
	}

	existingUserID := "existing-user-123"
	token, err := authMgr.GenerateToken(existingUserID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected user ID in context")
			http.Error(w, "no user ID", http.StatusInternalServerError)
			return
		}
		if userID != existingUserID {
			t.Errorf("expected user ID %q, got %q", existingUserID, userID)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userID))
	})

	middleware := WithCookie(authMgr)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "user_token",
		Value: token,
	})
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != existingUserID {
		t.Errorf("expected body %q, got %q", existingUserID, rr.Body.String())
	}
}

func TestWithCookie_InvalidToken(t *testing.T) {
	authMgr, err := auth.New(auth.Config{Secret: "testsecret", TokenTTL: 24 * time.Hour})
	if err != nil {
		t.Fatalf("failed to create auth manager: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected user ID in context")
			http.Error(w, "no user ID", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userID))
	})

	middleware := WithCookie(authMgr)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "user_token",
		Value: "invalid-token",
	})
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}
