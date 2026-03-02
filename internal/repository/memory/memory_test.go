package memory

import (
	"strings"
	"testing"
)

func TestRepository_Save(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "save valid URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "save another valid URL",
			url:     "https://google.com",
			wantErr: false,
		},
		{
			name:    "save URL with query params",
			url:     "https://example.com/path?query=value&another=param",
			wantErr: false,
		},
		{
			name:    "save empty URL",
			url:     "",
			wantErr: false,
		},
		{
			name:    "save URL with special characters",
			url:     "https://example.com/тест",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := New()

			shortID, err := repo.Save(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if shortID == "" {
					t.Error("save() returned empty short ID")
				}

				savedURL, exists := repo.data[shortID]
				if !exists {
					t.Error("URL was not saved in memory storage")
				}

				if savedURL != tt.url {
					t.Errorf("saved URL = %v, want %v", savedURL, tt.url)
				}
			}
		})
	}
}

func TestRepository_Get(t *testing.T) {
	repo := New()

	testURL := "https://example.com"
	shortID, err := repo.Save(testURL)
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
			url, ok := repo.Get(tt.id)

			if ok != tt.wantOk {
				t.Errorf("Get() ok = %v, want %v", ok, tt.wantOk)
			}

			if url != tt.wantURL {
				t.Errorf("Get() url = %v, want %v", url, tt.wantURL)
			}
		})
	}
}
