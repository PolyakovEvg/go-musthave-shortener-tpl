package memory

import (
	"strings"
	"testing"
)

var userID = "test_user"

func TestRepository_Save(t *testing.T) {
	tests := []struct {
		userID  string
		name    string
		url     string
		wantErr bool
	}{
		{
			userID:  userID,
			name:    "save valid URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			userID:  userID,
			name:    "save another valid URL",
			url:     "https://google.com",
			wantErr: false,
		},
		{
			userID:  userID,
			name:    "save URL with query params",
			url:     "https://example.com/path?query=value&another=param",
			wantErr: false,
		},
		{
			userID:  userID,
			name:    "save empty URL",
			url:     "",
			wantErr: false,
		},
		{
			userID:  userID,
			name:    "save URL with special characters",
			url:     "https://example.com/тест",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := New()

			shortID, err := repo.Save(tt.userID, tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if shortID == "" {
					t.Error("Save() returned empty short ID")
				}

				repo.mu.RLock()
				saved, exists := repo.data[shortID]
				repo.mu.RUnlock()

				if !exists {
					t.Error("URL was not saved in memory storage")
				} else if saved.OriginalURL != tt.url {
					t.Errorf("saved URL = %v, want %v", saved.OriginalURL, tt.url)
				}
			}
		})
	}
}

func TestRepository_Get(t *testing.T) {
	repo := New()

	testURL := "https://example.com"
	shortID, err := repo.Save(userID, testURL)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantURL string
		wantOk  bool
	}{
		{
			name:    "get existing URL",
			id:      shortID,
			wantURL: testURL,
			wantOk:  true,
		},
		{
			name:    "get non-existing URL",
			id:      "nonexistent",
			wantURL: "",
			wantOk:  false,
		},
		{
			name:    "get with empty ID",
			id:      "",
			wantURL: "",
			wantOk:  false,
		},
		{
			name:    "get with wrong case",
			id:      strings.ToUpper(shortID),
			wantURL: "",
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, ok := repo.Get(tt.id)

			if ok != tt.wantOk {
				t.Errorf("Get() ok = %v, want %v", ok, tt.wantOk)
			}

			gotURL := ""
			if ok && rec != nil {
				gotURL = rec.OriginalURL
			}

			if gotURL != tt.wantURL {
				t.Errorf("Get() url = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}
