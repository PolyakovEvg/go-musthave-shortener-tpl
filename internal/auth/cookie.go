// Package auth предоставляет аутентификацию пользователей через JWT токены в cookies.
// Используется для идентификации пользователей сервиса сокращения URL.
package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Ошибки аутентификации.
var (
	// ErrInvalidToken возвращается при невалидном или просроченном токене.
	ErrInvalidToken = errors.New("invalid token")
	// ErrNoCookie возвращается, когда cookie с токеном не найден.
	ErrNoCookie = errors.New("cookie not found")
)

// Config содержит параметры конфигурации менеджера аутентификации.
type Config struct {
	// Secret — секретный ключ для подписи JWT токенов.
	Secret string
	// TokenTTL — время жизни токена.
	TokenTTL time.Duration
	// CookieName — имя cookie для хранения токена.
	CookieName string
	// CookieMaxAge — максимальное время жизни cookie в секундах.
	CookieMaxAge int
	// HTTPOnly — флаг, запрещающий доступ к cookie из JavaScript.
	HTTPOnly bool
	// Secure — флаг, разрешающий передачу cookie только по HTTPS.
	Secure bool
}

// Manager управляет JWT токенами и cookies для аутентификации пользователей.
type Manager struct {
	secret       []byte
	tokenTTL     time.Duration
	cookieName   string
	cookieMaxAge int
	httpOnly     bool
	secure       bool
}

// Claims представляет claims JWT токена с идентификатором пользователя.
type Claims struct {
	// UserID — уникальный идентификатор пользователя.
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

// New создаёт новый менеджер аутентификации.
// Если некоторые параметры не указаны, используются значения по умолчанию.
func New(cfg Config) (*Manager, error) {
	if cfg.Secret == "" {
		return nil, errors.New("auth secret cannot be empty")
	}

	if cfg.TokenTTL == 0 {
		cfg.TokenTTL = 24 * time.Hour
	}
	if cfg.CookieName == "" {
		cfg.CookieName = "user_token"
	}
	if cfg.CookieMaxAge == 0 {
		cfg.CookieMaxAge = 3600 * 24 * 30
	}

	return &Manager{
		secret:       []byte(cfg.Secret),
		tokenTTL:     cfg.TokenTTL,
		cookieName:   cfg.CookieName,
		cookieMaxAge: cfg.CookieMaxAge,
		httpOnly:     cfg.HTTPOnly,
		secure:       cfg.Secure,
	}, nil
}

// GenerateToken генерирует JWT токен для указанного пользователя.
// Возвращает подписанный токен или ошибку.
func (m *Manager) GenerateToken(userID string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(m.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseToken парсит JWT токен и возвращает идентификатор пользователя.
// Возвращает ErrInvalidToken, если токен невалиден или просрочен.
func (m *Manager) ParseToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	if claims.UserID == "" {
		return "", ErrInvalidToken
	}

	return claims.UserID, nil
}

// SetCookie устанавливает cookie с JWT токеном для указанного пользователя.
// Токен автоматически генерируется и подписывается.
func (m *Manager) SetCookie(w http.ResponseWriter, userID string) error {
	token, err := m.GenerateToken(userID)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: m.httpOnly,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   m.cookieMaxAge,
		Expires:  time.Now().Add(m.tokenTTL),
	})

	return nil
}

// GetUserID извлекает идентификатор пользователя из cookie в запросе.
// Возвращает ErrNoCookie, если cookie не найден, или ErrInvalidToken, если токен невалиден.
func (m *Manager) GetUserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return "", ErrNoCookie
	}

	return m.ParseToken(cookie.Value)
}

// GetOrCreateUserID возвращает идентификатор пользователя из cookie или создаёт нового.
// Если cookie отсутствует или токен невалиден, генерируется новый UUID и устанавливается cookie.
func (m *Manager) GetOrCreateUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	userID, err := m.GetUserID(r)
	if err == nil {
		return userID, nil
	}

	if errors.Is(err, ErrNoCookie) {
		userID = uuid.New().String()
		if err := m.SetCookie(w, userID); err != nil {
			return "", err
		}
		return userID, nil
	}

	if errors.Is(err, ErrInvalidToken) {
		userID = uuid.New().String()
		if err := m.SetCookie(w, userID); err != nil {
			return "", err
		}
		return userID, nil
	}

	return "", err
}
