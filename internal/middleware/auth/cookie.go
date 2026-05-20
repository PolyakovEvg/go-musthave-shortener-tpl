// Package auth предоставляет middleware для аутентификации пользователей.
// Извлекает или создаёт идентификатор пользователя из cookie и добавляет его в контекст.
package auth

import (
	"context"
	"log"
	"net/http"

	"PolyakovEvg/go-musthave-shortener-tpl/internal/auth"
)

// contextKey — тип для ключей контекста.
type contextKey struct{}

// userIDKey — ключ для хранения идентификатора пользователя в контексте.
var userIDKey = contextKey{}

// UserIDFromContext извлекает идентификатор пользователя из контекста запроса.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

// WithCookie возвращает middleware, который извлекает или создаёт идентификатор пользователя.
// Идентификатор сохраняется в контексте запроса и может быть получен через UserIDFromContext.
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
