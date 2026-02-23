package repository

import "PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"

type Storage struct {
	data map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
	}
}

func (s *Storage) Save(url string) (string, error) {
	safeStr, err := randstr.GenerateRandomStringURLSafe(8)

	if err != nil {
		return "", err
	}

	s.data[safeStr] = url
	return safeStr, nil
}

func (s *Storage) Get(id string) (string, bool) {
	url, exists := s.data[id]
	return url, exists
}
