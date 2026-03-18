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
		return nil, fmt.Errorf("Error loading repository: %w", err)
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
		return fmt.Errorf("Error unmarshaling JSON from %s: %v", r.filePath, err)
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

func (r *FileRepository) Save(originalURL string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	shortID, err := randstr.GenerateRandomStringURLSafe(8)
	if err != nil {
		return "", fmt.Errorf("Error generating random string: %v", err)
	}

	r.counter++
	uuid := fmt.Sprintf("%d", r.counter)

	rec := model.FileRecord{
		UUID:        uuid,
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	r.data[shortID] = rec

	if err := r.flush(); err != nil {
		return "", fmt.Errorf("Error flushing data to file: %v", err)
	}

	return shortID, nil
}

func (r *FileRepository) SaveBatch(batch []model.BatchRequest) ([]model.BatchResponse, error) {
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
			return nil, fmt.Errorf("Error generating random string: %v", err)
		}

		r.counter++
		uuid := fmt.Sprintf("%d", r.counter)

		rec := model.FileRecord{
			UUID:        uuid,
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

func (r *FileRepository) Get(shortURL string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rec, ok := r.data[shortURL]
	if !ok {
		return "", false
	}

	return rec.OriginalURL, true
}

func (r *FileRepository) flush() error {

	records := make([]model.FileRecord, 0, len(r.data))
	for _, rec := range r.data {
		records = append(records, rec)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshaling records to JSON: %v", err)
	}

	dir := filepath.Dir(r.filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("Error creating directory %s: %v", dir, err)
		}
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("write file %s: %w", r.filePath, err)
	}

	return nil
}
