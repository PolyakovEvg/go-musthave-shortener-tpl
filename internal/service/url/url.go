package url

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"errors"
	"fmt"
)

type URLService struct {
	repo    repository.Repository
	baseURL string
}

func NewURLService(repo repository.Repository, baseURL string) *URLService {
	return &URLService{
		repo:    repo,
		baseURL: baseURL,
	}
}

func (s *URLService) SaveShorten(userID, original string) (string, error) {
	if original == "" {
		return "", errors.New("url is empty")
	}

	id, err := s.repo.Save(userID, original)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return s.baseURL + id, err
		}
		return "", err
	}

	return s.baseURL + id, nil
}

func (s *URLService) GetOriginal(id string) (*model.URL, error) {
	rec, ok := s.repo.Get(id)

	if !ok {
		return nil, errors.New("not found original URL")
	}
	return rec, nil
}

func (s *URLService) SaveBatch(userID string, batch []model.BatchRequest) ([]model.BatchResponse, error) {
	if len(batch) == 0 {
		return nil, errors.New("empty batch")
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

func (s *URLService) GetUserURLs(userID string) ([]model.URL, error) {
	if userID == "" {
		return nil, errors.New("empty user id")
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

func (s *URLService) PingRepository() error {
	err := s.repo.Ping()

	if err != nil {
		return fmt.Errorf("failed to ping repository: %w", err)
	}
	return nil
}
