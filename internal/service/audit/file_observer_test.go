package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFileObserver(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		errType  error
	}{
		{
			name:     "empty file path",
			filePath: "",
			wantErr:  true,
			errType:  ErrEmptyFilePath,
		},
		{
			name:     "valid file path",
			filePath: filepath.Join(t.TempDir(), "audit.log"),
			wantErr:  false,
		},
		{
			name:     "path with non-existent directory",
			filePath: filepath.Join(t.TempDir(), "subdir", "audit.log"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observer, err := NewFileObserver(tt.filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewFileObserver() expected error, got nil")
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewFileObserver() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("NewFileObserver() unexpected error: %v", err)
			}

			if observer == nil {
				t.Errorf("NewFileObserver() returned nil observer")
			}

			if observer.file == nil {
				t.Errorf("NewFileObserver() file is nil")
			}

			if _, err := os.Stat(tt.filePath); os.IsNotExist(err) {
				t.Errorf("file %s was not created", tt.filePath)
			}

			observer.Close()
		})
	}
}

func TestFileObserver_Send(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "audit.log")

	observer, err := NewFileObserver(filePath)
	if err != nil {
		t.Fatalf("Failed to create observer: %v", err)
	}
	defer observer.Close()

	tests := []struct {
		name  string
		event AuditEvent
	}{
		{
			name: "shorten event",
			event: AuditEvent{
				Ts:     time.Now().Unix(),
				Action: "shorten",
				UserID: "user123",
				URL:    "https://example.com/long/url",
			},
		},
		{
			name: "follow event without user",
			event: AuditEvent{
				Ts:     time.Now().Unix(),
				Action: "follow",
				URL:    "https://example.com/another",
			},
		},
		{
			name: "event with empty fields",
			event: AuditEvent{
				Ts:     time.Now().Unix(),
				Action: "shorten",
				URL:    "https://test.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := observer.Send(tt.event)
			if err != nil {
				t.Errorf("Send() error = %v", err)
			}
		})
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}

	if lines != len(tests) {
		t.Errorf("Expected %d lines, got %d", len(tests), lines)
	}
}
