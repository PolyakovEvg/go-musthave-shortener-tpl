package audit

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var ErrEmptyFilePath = errors.New("file path is empty")

type FileObserver struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileObserver(filePath string) (*FileObserver, error) {
	if filePath == "" {
		return nil, ErrEmptyFilePath
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileObserver{file: file}, nil
}

func (f *FileObserver) Send(event AuditEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = f.file.Write(data)
	return err
}

func (f *FileObserver) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}
