package url

import (
	"errors"
)

type Repository interface {
	Save(string) (string, error)
	Get(string) (string, bool)
}

type URLService struct {
	repo    Repository
	baseURL string
}

func NewURLService(repo Repository, baseURL string) *URLService {
	return &URLService{
		repo:    repo,
		baseURL: baseURL,
	}
}

func (s *URLService) SaveShorten(original string) (string, error) {
	if original == "" {
		return "", errors.New("url is empty")
	}

	id, err := s.repo.Save(original)
	if err != nil {
		return "", err
	}

	return s.baseURL + id, nil
}

func (s *URLService) GetOriginal(id string) (string, error) {
	url, ok := s.repo.Get(id)

	if !ok {
		return "", errors.New("not found original URL")
	}
	return url, nil
}
