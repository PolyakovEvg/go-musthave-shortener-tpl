package memory

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"
	"sync"
)

type MemoryRepository struct {
	data  map[string]model.URL
	byURL map[string]string
	mu    sync.RWMutex
}

func New() *MemoryRepository {
	return &MemoryRepository{
		data:  make(map[string]model.URL),
		byURL: make(map[string]string),
	}
}

func (mr *MemoryRepository) Ping() error {
	return nil
}

func (mr *MemoryRepository) Save(userID, original string) (string, error) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if short, exists := mr.byURL[original]; exists {
		return short, nil
	}

	shortID, err := randstr.GenerateRandomStringURLSafe(8)
	if err != nil {
		return "", err
	}

	mr.data[shortID] = model.URL{
		ShortURL:    shortID,
		OriginalURL: original,
		UserID:      userID,
	}

	mr.byURL[original] = shortID

	return shortID, nil
}

func (mr *MemoryRepository) Get(shortURL string) (*model.URL, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	rec, exists := mr.data[shortURL]
	if !exists {
		return nil, false
	}

	return &rec, true
}

func (mr *MemoryRepository) SaveBatch(userID string, batch []model.BatchRequest) ([]model.BatchResponse, error) {
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

		mr.data[shortID] = model.URL{
			ShortURL:    shortID,
			OriginalURL: req.OriginalURL,
			UserID:      userID,
		}

		mr.byURL[req.OriginalURL] = shortID

		responses = append(responses, model.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortID,
		})
	}

	return responses, nil
}

func (mr *MemoryRepository) GetByUser(userID string) ([]model.URL, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	result := make([]model.URL, 0)

	for _, u := range mr.data {
		if u.UserID == userID {
			result = append(result, u)
		}
	}

	return result, nil
}

func (mr *MemoryRepository) MarkDeleted(userID string, shorts []string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	for _, short := range shorts {
		rec, ok := mr.data[short]
		if !ok {
			continue
		}
		if rec.UserID != userID {
			continue
		}
		rec.IsDeleted = true
		mr.data[short] = rec
	}
	return nil
}
