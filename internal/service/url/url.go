// Package url предоставляет сервис для работы с сокращёнными URL.
// Содержит бизнес-логику создания и получения коротких URL.
package url

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/apperrors"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"errors"
	"fmt"
)

// URLService предоставляет методы для работы с сокращёнными URL.
type URLService struct {
	repo    repository.Repository
	baseURL string
}

// NewURLService создаёт новый сервис URL с указанным хранилищем и базовым URL.
func NewURLService(repo repository.Repository, baseURL string) *URLService {
	return &URLService{
		repo:    repo,
		baseURL: baseURL,
	}
}

// SaveShorten сохраняет оригинальный URL и возвращает сокращённый.
// Если URL уже существует, возвращает существующий короткий URL с ошибкой apperrors.ErrURLConflict.
func (s *URLService) SaveShorten(userID, original string) (string, error) {
	if original == "" {
		return "", apperrors.ErrEmptyURL
	}

	id, err := s.repo.Save(userID, original)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return s.baseURL + id, apperrors.ErrURLConflict
		}
		return "", err
	}

	return s.baseURL + id, nil
}

// GetOriginal возвращает оригинальный URL по короткому идентификатору.
// Возвращает ошибку apperrors.ErrURLNotFound, если URL не найден.
func (s *URLService) GetOriginal(id string) (*model.URL, error) {
	rec, ok := s.repo.Get(id)

	if !ok {
		return nil, apperrors.ErrURLNotFound
	}
	return rec, nil
}

// SaveBatch сохраняет пакет URL и возвращает сокращённые URL.
// Принимает список запросов и возвращает список ответов с короткими URL.
func (s *URLService) SaveBatch(userID string, batch []model.BatchRequest) ([]model.BatchResponse, error) {
	if len(batch) == 0 {
		return nil, apperrors.ErrEmptyBatch
	}

	responses, err := s.repo.SaveBatch(userID, batch)
	if err != nil {
		return nil, err
	}

	for i := range responses {
		responses[i].ShortURL = s.baseURL + responses[i].ShortURL
	}

	return responses, nil
}

// GetUserURLs возвращает все URL, созданные указанным пользователем.
// Возвращает ошибку apperrors.ErrEmptyUserID, если userID пуст.
func (s *URLService) GetUserURLs(userID string) ([]model.URL, error) {
	if userID == "" {
		return nil, apperrors.ErrEmptyUserID
	}

	urls, err := s.repo.GetByUser(userID)
	if err != nil {
		return nil, err
	}

	for i := range urls {
		urls[i].ShortURL = s.baseURL + urls[i].ShortURL
	}

	return urls, nil
}

// PingRepository проверяет соединение с хранилищем.
// Используется для health-check сервиса.
func (s *URLService) PingRepository() error {
	err := s.repo.Ping()

	if err != nil {
		return fmt.Errorf("failed to ping repository: %w", err)
	}
	return nil
}
