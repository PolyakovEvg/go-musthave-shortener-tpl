package memory

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"
	"sync"
)

type MemoryRepository struct {
	data map[string]string
	mu   sync.RWMutex
}

func New() *MemoryRepository {
	return &MemoryRepository{
		data: make(map[string]string),
	}
}

func (mr *MemoryRepository) Save(url string) (string, error) {
	safeStr, err := randstr.GenerateRandomStringURLSafe(8)

	if err != nil {
		return "", err
	}

	mr.mu.Lock()
	defer mr.mu.Unlock()

	mr.data[safeStr] = url
	return safeStr, nil
}

func (mr *MemoryRepository) Get(id string) (string, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	url, exists := mr.data[id]
	return url, exists
}
