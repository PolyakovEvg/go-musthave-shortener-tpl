package memory

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"
	"sync"
)

type MemoryRepository struct {
	data map[string]string
	mu   sync.Mutex
}

func New() *MemoryRepository {
	return &MemoryRepository{
		data: make(map[string]string),
	}
}

func (s *MemoryRepository) Save(url string) (string, error) {
	safeStr, err := randstr.GenerateRandomStringURLSafe(8)

	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[safeStr] = url
	return safeStr, nil
}

func (s *MemoryRepository) Get(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, exists := s.data[id]
	return url, exists
}
