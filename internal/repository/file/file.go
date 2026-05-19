package file

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FileRepository struct {
	filePath string
	data     map[string]model.FileRecord
	counter  int
	mu       sync.RWMutex
}

func New(filePath string) (*FileRepository, error) {
	repo := &FileRepository{
		filePath: filePath,
		data:     make(map[string]model.FileRecord),
	}

	if err := repo.load(); err != nil {
		return nil, fmt.Errorf("error loading repository: %w", err)
	}

	return repo, nil
}

func (r *FileRepository) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fileBytes, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error reading file %s: %v", r.filePath, err)
	}

	if len(fileBytes) == 0 {
		return nil
	}

	var records []model.FileRecord
	if err := json.Unmarshal(fileBytes, &records); err != nil {
		return fmt.Errorf("error unmarshaling JSON from %s: %v", r.filePath, err)
	}

	for _, rec := range records {
		r.data[rec.ShortURL] = rec

		var id int
		fmt.Sscanf(rec.UUID, "%d", &id)
		if id > r.counter {
			r.counter = id
		}
	}
	return nil
}

func (r *FileRepository) Ping() error {
	return nil
}

func (r *FileRepository) Save(userID, originalURL string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	shortID, err := randstr.GenerateRandomStringURLSafe(8)
	if err != nil {
		return "", fmt.Errorf("error generating random string: %v", err)
	}

	r.counter++
	uuid := fmt.Sprintf("%d", r.counter)

	rec := model.FileRecord{
		UUID:        uuid,
		UserID:      userID,
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	r.data[shortID] = rec

	if err := r.flush(); err != nil {
		return "", fmt.Errorf("error flushing data to file: %v", err)
	}

	return shortID, nil
}

func (r *FileRepository) SaveBatch(userID string, batch []model.BatchRequest) ([]model.BatchResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	responses := make([]model.BatchResponse, 0, len(batch))
	newRecords := make([]model.FileRecord, 0, len(batch))
	urlToShort := make(map[string]string, len(r.data))

	for _, rec := range r.data {
		urlToShort[rec.OriginalURL] = rec.ShortURL
	}

	for _, req := range batch {
		if existingShort, exists := urlToShort[req.OriginalURL]; exists {
			responses = append(responses, model.BatchResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      existingShort,
			})
			continue
		}

		shortID, err := randstr.GenerateRandomStringURLSafe(8)
		if err != nil {
			return nil, fmt.Errorf("error generating random string: %v", err)
		}

		r.counter++
		uuid := fmt.Sprintf("%d", r.counter)

		rec := model.FileRecord{
			UUID:        uuid,
			UserID:      userID,
			ShortURL:    shortID,
			OriginalURL: req.OriginalURL,
		}

		r.data[shortID] = rec
		urlToShort[req.OriginalURL] = shortID
		newRecords = append(newRecords, rec)

		responses = append(responses, model.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortID,
		})
	}

	if len(newRecords) > 0 {
		if err := r.flush(); err != nil {
			return nil, fmt.Errorf("Error flushing data to file: %v", err)
		}
	}

	return responses, nil
}

func (r *FileRepository) Get(shortURL string) (*model.URL, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rec, ok := r.data[shortURL]
	if !ok {
		return nil, false
	}

	return &model.URL{
		ShortURL:    rec.ShortURL,
		OriginalURL: rec.OriginalURL,
		UserID:      rec.UserID,
		IsDeleted:   rec.IsDeleted,
	}, true
}

func (r *FileRepository) GetByUser(userID string) ([]model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]model.URL, 0)

	for _, rec := range r.data {
		if rec.UserID == userID {
			result = append(result, model.URL{
				ShortURL:    rec.ShortURL,
				OriginalURL: rec.OriginalURL,
				UserID:      rec.UserID,
			})
		}
	}

	return result, nil
}

func (r *FileRepository) MarkDeleted(userID string, shorts []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	changed := false

	for _, s := range shorts {
		rec, ok := r.data[s]
		if !ok {
			continue
		}

		if rec.UserID != userID {
			continue
		}

		if !rec.IsDeleted {
			rec.IsDeleted = true
			r.data[s] = rec
			changed = true
		}
	}

	if changed {
		if err := r.flush(); err != nil {
			return err
		}
	}

	return nil
}

func (r *FileRepository) flush() error {

	records := make([]model.FileRecord, 0, len(r.data))
	for _, rec := range r.data {
		records = append(records, rec)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling records to JSON: %v", err)
	}

	dir := filepath.Dir(r.filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %v", dir, err)
		}
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("write file %s: %w", r.filePath, err)
	}

	return nil
}
