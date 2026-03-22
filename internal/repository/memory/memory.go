package memory

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"
	"sync"
)

type MemoryRepository struct {
	byShort map[string]string
	byURL   map[string]string
	mu      sync.RWMutex
}

func New() *MemoryRepository {
	return &MemoryRepository{
		byShort: make(map[string]string),
		byURL:   make(map[string]string),
	}
}

func (mem *MemoryRepository) Ping() error {
	return nil
}

func (mr *MemoryRepository) Save(url string) (string, error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if short, exists := mr.byURL[url]; exists {
		return short, nil
	}

	shortID, err := randstr.GenerateRandomStringURLSafe(8)
	if err != nil {
		return "", err
	}

	mr.byShort[shortID] = url
	mr.byURL[url] = shortID

	return shortID, nil
}

func (mr *MemoryRepository) Get(id string) (string, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	url, exists := mr.byShort[id]
	return url, exists
}

func (mr *MemoryRepository) SaveBatch(batch []model.BatchRequest) ([]model.BatchResponse, error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	responses := make([]model.BatchResponse, 0, len(batch))

	for _, req := range batch {
		if short, exists := mr.byURL[req.OriginalURL]; exists {
			responses = append(responses, model.BatchResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      short,
			})
			continue
		}

		shortID, err := randstr.GenerateRandomStringURLSafe(8)
		if err != nil {
			return nil, err
		}

		mr.byShort[shortID] = req.OriginalURL
		mr.byURL[req.OriginalURL] = shortID

		responses = append(responses, model.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortID,
		})
	}

	return responses, nil
}
