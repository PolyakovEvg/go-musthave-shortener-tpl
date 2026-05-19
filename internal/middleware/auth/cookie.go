package auth

import (
	"context"
	"log"
	"net/http"

	"PolyakovEvg/go-musthave-shortener-tpl/internal/auth"
)

type contextKey struct{}

var userIDKey = contextKey{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

func WithCookie(m *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := m.GetOrCreateUserID(w, r)
			if err != nil {
				log.Printf("auth error: %v", err)
				http.Error(w, "authentication failed", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
