package file

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"
	"encoding/json"
	"fmt"
	"log"
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
		log.Printf("Error loading repository: %v", err)
		return nil, err
	}

	log.Printf("Repository initialized successfully. Loaded %d records", len(repo.data))
	return repo, nil
}

func (r *FileRepository) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("Loading data from file: %s", r.filePath)

	fileBytes, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File %s does not exist, starting with empty repository", r.filePath)
			return nil
		}
		log.Printf("Error reading file %s: %v", r.filePath, err)
		return err
	}

	if len(fileBytes) == 0 {
		log.Printf("File %s is empty, starting with empty repository", r.filePath)
		return nil
	}

	log.Printf("Read %d bytes from file", len(fileBytes))

	var records []model.FileRecord
	if err := json.Unmarshal(fileBytes, &records); err != nil {
		log.Printf("Error unmarshaling JSON from %s: %v", r.filePath, err)
		return err
	}

	log.Printf("Successfully unmarshaled %d records", len(records))

	for _, rec := range records {
		r.data[rec.ShortURL] = rec

		var id int
		fmt.Sscanf(rec.UUID, "%d", &id)
		if id > r.counter {
			r.counter = id
		}
	}

	log.Printf("Loaded %d records, counter is now %d", len(r.data), r.counter)
	return nil
}

func (r *FileRepository) Save(originalURL string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("Saving original URL: %s", originalURL)

	shortID, err := randstr.GenerateRandomStringURLSafe(8)
	if err != nil {
		log.Printf("Error generating random string: %v", err)
		return "", err
	}

	r.counter++
	uuid := fmt.Sprintf("%d", r.counter)

	rec := model.FileRecord{
		UUID:        uuid,
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	r.data[shortID] = rec
	log.Printf("Created record: UUID=%s, ShortURL=%s, OriginalURL=%s", uuid, shortID, originalURL)

	if err := r.flush(); err != nil {
		log.Printf("Error flushing data to file: %v", err)
		return "", err
	}

	log.Printf("Successfully saved URL with short ID: %s", shortID)
	return shortID, nil
}

func (r *FileRepository) SaveBatch(batch []model.BatchRequest) ([]model.BatchResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	responses := make([]model.BatchResponse, 0, len(batch))
	newRecords := make([]model.FileRecord, 0, len(batch))

	for _, req := range batch {
		var existingShort string
		var found bool

		for _, rec := range r.data {
			if rec.OriginalURL == req.OriginalURL {
				existingShort = rec.ShortURL
				found = true
				break
			}
		}

		if found {
			responses = append(responses, model.BatchResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      existingShort,
			})
			continue
		}

		shortID, err := randstr.GenerateRandomStringURLSafe(8)
		if err != nil {
			log.Printf("Error generating random string: %v", err)
			return nil, err
		}

		r.counter++
		uuid := fmt.Sprintf("%d", r.counter)

		rec := model.FileRecord{
			UUID:        uuid,
			ShortURL:    shortID,
			OriginalURL: req.OriginalURL,
		}

		r.data[shortID] = rec
		newRecords = append(newRecords, rec)

		responses = append(responses, model.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortID,
		})

		log.Printf("Created record: UUID=%s, ShortURL=%s, OriginalURL=%s",
			uuid, shortID, req.OriginalURL)
	}

	if len(newRecords) > 0 {
		log.Printf("Flushing %d new records to file", len(newRecords))
		if err := r.flush(); err != nil {
			log.Printf("Error flushing data to file: %v", err)
			return nil, err
		}
	}

	log.Printf("Successfully saved batch of %d URLs", len(batch))
	return responses, nil
}

func (r *FileRepository) Get(shortURL string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	log.Printf("Getting original URL for short URL: %s", shortURL)

	rec, ok := r.data[shortURL]
	if !ok {
		log.Printf("Short URL %s not found", shortURL)
		return "", false
	}

	log.Printf("Found original URL: %s for short URL: %s", rec.OriginalURL, shortURL)
	return rec.OriginalURL, true
}

func (r *FileRepository) flush() error {
	log.Printf("Flushing data to file: %s", r.filePath)

	records := make([]model.FileRecord, 0, len(r.data))
	for _, rec := range r.data {
		records = append(records, rec)
	}

	log.Printf("Preparing to write %d records to file", len(records))

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		log.Printf("Error marshaling records to JSON: %v", err)
		return err
	}

	log.Printf("Marshaled %d bytes of JSON data", len(data))

	dir := filepath.Dir(r.filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Error creating directory %s: %v", dir, err)
			return fmt.Errorf("failed to create directory: %w", err)
		}
		log.Printf("Created directory: %s", dir)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		log.Printf("Error writing to file %s: %v", r.filePath, err)

		if os.IsPermission(err) {
			log.Printf("Permission denied. Check write permissions for %s", r.filePath)
		}
		if os.IsNotExist(err) {
			log.Printf("Parent directory might not exist")
		}

		return err
	}

	if fileInfo, err := os.Stat(r.filePath); err == nil {
		log.Printf("Successfully wrote to file: %s (size: %d bytes)", r.filePath, fileInfo.Size())
	} else {
		log.Printf("Warning: File was written but stat failed: %v", err)
	}

	return nil
}
