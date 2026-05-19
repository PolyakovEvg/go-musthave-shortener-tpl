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
	Secure       bool
}

type Manager struct {
	secret       []byte
	tokenTTL     time.Duration
	cookieName   string
	cookieMaxAge int
	httpOnly     bool
	secure       bool
}

type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

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
