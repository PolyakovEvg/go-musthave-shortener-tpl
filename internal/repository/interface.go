// Package repository предоставляет интерфейс для хранения URL.
package repository

import "PolyakovEvg/go-musthave-shortener-tpl/internal/model"

//go:generate mockgen -source=interface.go -destination=mocks/mock_repository.go -package=mocks Repository

// Repository определяет интерфейс для хранения и получения сокращённых URL.
// Реализации должны быть потокобезопасными.
type Repository interface {
	// Save сохраняет оригинальный URL и возвращает короткий идентификатор.
	// Если URL уже существует, возвращает существующий короткий идентификатор без ошибки.
	Save(userID, originalURL string) (shortID string, err error)

	// SaveBatch сохраняет несколько URL за одну операцию.
	// Возвращает список ответов с короткими идентификаторами в том же порядке.
	SaveBatch(userID string, batch []model.BatchRequest) ([]model.BatchResponse, error)

	// Ping проверяет доступность хранилища.
	// Возвращает nil, если хранилище доступно.
	Ping() error

	// Get возвращает URL по короткому идентификатору.
	// Возвращает false, если URL не найден.
	Get(shortID string) (*model.URL, bool)

	// GetByUser возвращает все URL, созданные указанным пользователем.
	// Возвращает пустой слайс, если у пользователя нет URL.
	GetByUser(userID string) ([]model.URL, error)

	// MarkDeleted помечает указанные URL как удалённые.
	// Удалять может только владелец URL.
	MarkDeleted(userID string, shortIDs []string) error

	// Close закрывает соединение с хранилищем и сохраняет несохранённые данные.
	Close() error
}
