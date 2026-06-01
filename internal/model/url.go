// Package model содержит структуры данных для сервиса сокращения URL.
package model

// URL представляет запись о сокращённом URL.
type URL struct {
	// ShortURL — короткий идентификатор URL.
	ShortURL string
	// OriginalURL — оригинальный (длинный) URL.
	OriginalURL string
	// UserID — идентификатор пользователя, создавшего короткий URL.
	UserID string
	// IsDeleted — флаг, указывающий, что URL был удалён.
	IsDeleted bool
}
