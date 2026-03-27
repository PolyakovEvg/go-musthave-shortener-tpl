package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrNoCookie     = errors.New("cookie not found")
)

type Config struct {
	Secret       string
	TokenTTL     time.Duration
	CookieName   string
	CookieMaxAge int
	HTTPOnly     bool
}

type Manager struct {
	secret       []byte
	tokenTTL     time.Duration
	cookieName   string
	cookieMaxAge int
	httpOnly     bool
}

func New(cfg Config) *Manager {
	if cfg.Secret == "" {
		panic("auth secret cannot be empty")
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
	}
}

func (m *Manager) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(m.tokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) ParseToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", ErrInvalidToken
	}

	return sub, nil
}

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
		MaxAge:   m.cookieMaxAge,
	})

	return nil
}

func (m *Manager) GetUserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return "", ErrNoCookie
	}

	return m.ParseToken(cookie.Value)
}

func (m *Manager) GetOrCreateUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	userID, err := m.GetUserID(r)
	if err == nil {
		return userID, nil
	}

	if errors.Is(err, ErrNoCookie) || errors.Is(err, ErrInvalidToken) {
		userID = uuid.New().String()
		if err := m.SetCookie(w, userID); err != nil {
			return "", err
		}
		return userID, nil
	}

	return "", err
}

func (m *Manager) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: m.httpOnly,
	})
}
