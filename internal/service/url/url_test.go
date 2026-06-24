package url

import (
	"errors"
	"testing"

	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/memory"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/mocks"

	"go.uber.org/mock/gomock"
)

func TestURLService_SaveShorten(t *testing.T) {
	repo := memory.New()
	service := NewURLService(repo, "http://localhost:8080/")

	tests := []struct {
		name        string
		userID      string
		originalURL string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid URL",
			userID:      "user1",
			originalURL: "https://example.com",
			wantErr:     false,
		},
		{
			name:        "empty URL",
			userID:      "user1",
			originalURL: "",
			wantErr:     true,
			errMsg:      "url is empty",
		},
		{
			name:        "another valid URL",
			userID:      "user2",
			originalURL: "https://github.com",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortURL, err := service.SaveShorten(tt.userID, tt.originalURL)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if shortURL == "" {
				t.Error("expected short URL, got empty string")
			}
		})
	}
}

func TestURLService_SaveShorten_SameURL(t *testing.T) {
	repo := memory.New()
	service := NewURLService(repo, "http://localhost:8080/")

	originalURL := "https://example.com"
	shortURL1, err := service.SaveShorten("user1", originalURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	shortURL2, err := service.SaveShorten("user2", originalURL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if shortURL2 != shortURL1 {
		t.Errorf("expected same short URL, got %s vs %s", shortURL1, shortURL2)
	}
}

func TestURLService_SaveShorten_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().
		Save("user1", "https://example.com").
		Return("abc123", repository.ErrConflict)

	service := NewURLService(mockRepo, "http://localhost:8080/")

	shortURL, err := service.SaveShorten("user1", "https://example.com")
	if err == nil {
		t.Error("expected conflict error, got nil")
		return
	}

	if !errors.Is(err, repository.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}

	if shortURL != "http://localhost:8080/abc123" {
		t.Errorf("expected short URL with base prefix, got %s", shortURL)
	}
}

func TestURLService_GetOriginal(t *testing.T) {
	repo := memory.New()
	service := NewURLService(repo, "http://localhost:8080/")

	originalURL := "https://example.com"
	shortURL, err := service.SaveShorten("user1", originalURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	shortID := shortURL[len("http://localhost:8080/"):]

	rec, err := service.GetOriginal(shortID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.OriginalURL != originalURL {
		t.Errorf("expected original URL %q, got %q", originalURL, rec.OriginalURL)
	}
}

func TestURLService_GetOriginal_NotFound(t *testing.T) {
	repo := memory.New()
	service := NewURLService(repo, "http://localhost:8080/")

	_, err := service.GetOriginal("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent URL, got nil")
		return
	}

	if err.Error() != "not found original URL" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestURLService_SaveBatch(t *testing.T) {
	repo := memory.New()
	service := NewURLService(repo, "http://localhost:8080/")

	batch := []model.BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
		{CorrelationID: "3", OriginalURL: "https://example3.com"},
	}

	responses, err := service.SaveBatch("user1", batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(responses) != len(batch) {
		t.Errorf("expected %d responses, got %d", len(batch), len(responses))
	}

	for i, resp := range responses {
		if resp.CorrelationID != batch[i].CorrelationID {
			t.Errorf("expected correlation ID %q, got %q", batch[i].CorrelationID, resp.CorrelationID)
		}

		if resp.ShortURL == "" {
			t.Error("expected short URL, got empty string")
		}
	}
}

func TestURLService_SaveBatch_Empty(t *testing.T) {
	repo := memory.New()
	service := NewURLService(repo, "http://localhost:8080/")

	_, err := service.SaveBatch("user1", []model.BatchRequest{})
	if err == nil {
		t.Error("expected error for empty batch, got nil")
		return
	}

	if err.Error() != "empty batch" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestURLService_GetUserURLs(t *testing.T) {
	repo := memory.New()
	service := NewURLService(repo, "http://localhost:8080/")

	userID := "user1"
	urls := []string{
		"https://example1.com",
		"https://example2.com",
		"https://example3.com",
	}

	for _, u := range urls {
		_, err := service.SaveShorten(userID, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	result, err := service.GetUserURLs(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(urls) {
		t.Errorf("expected %d URLs, got %d", len(urls), len(result))
	}
}

func TestURLService_GetUserURLs_EmptyUserID(t *testing.T) {
	repo := memory.New()
	service := NewURLService(repo, "http://localhost:8080/")

	_, err := service.GetUserURLs("")
	if err == nil {
		t.Error("expected error for empty user ID, got nil")
		return
	}

	if err.Error() != "empty user id" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestURLService_PingRepository(t *testing.T) {
	repo := memory.New()
	service := NewURLService(repo, "http://localhost:8080/")

	err := service.PingRepository()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestURLService_SaveBatch_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().
		SaveBatch("user1", gomock.Any()).
		Return(nil, errors.New("batch error"))

	service := NewURLService(mockRepo, "http://localhost:8080/")

	batch := []model.BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
	}

	_, err := service.SaveBatch("user1", batch)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestURLService_GetUserURLs_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().
		GetByUser("user1").
		Return(nil, errors.New("get by user error"))

	service := NewURLService(mockRepo, "http://localhost:8080/")

	_, err := service.GetUserURLs("user1")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestURLService_PingRepository_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().
		Ping().
		Return(errors.New("ping error"))

	service := NewURLService(mockRepo, "http://localhost:8080/")

	err := service.PingRepository()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestURLService_Save_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().
		Save("user1", "https://example.com").
		Return("", errors.New("save error"))

	service := NewURLService(mockRepo, "http://localhost:8080/")

	_, err := service.SaveShorten("user1", "https://example.com")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
