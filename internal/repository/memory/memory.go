package memory

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
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
	mr.mu.RLock()
	for short, existingURL := range mr.data {
		if existingURL == url {
			mr.mu.RUnlock()
			return short, nil
		}
	}
	mr.mu.RUnlock()

	shortID, err := randstr.GenerateRandomStringURLSafe(8)
	if err != nil {
		return "", err
	}

	mr.mu.Lock()
	defer mr.mu.Unlock()

	for short, existingURL := range mr.data {
		if existingURL == url {
			return short, nil
		}
	}

	mr.data[shortID] = url
	return shortID, nil
}

func (mr *MemoryRepository) Get(id string) (string, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	url, exists := mr.data[id]
	return url, exists
}

func (mr *MemoryRepository) SaveBatch(batch []model.BatchRequest) ([]model.BatchResponse, error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	responses := make([]model.BatchResponse, 0, len(batch))

	for _, req := range batch {
		found := false
		for short, existingURL := range mr.data {
			if existingURL == req.OriginalURL {
				responses = append(responses, model.BatchResponse{
					CorrelationID: req.CorrelationID,
					ShortURL:      short,
				})
				found = true
				break
			}
		}

		if found {
			continue
		}

		shortID, err := randstr.GenerateRandomStringURLSafe(8)
		if err != nil {
			return nil, err
		}

		mr.data[shortID] = req.OriginalURL

		responses = append(responses, model.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortID,
		})
	}

	return responses, nil
}
