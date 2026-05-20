package memory

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"strconv"
	"testing"
)

var userID = "test_user"

func TestMemoryRepository_Save(t *testing.T) {
	repo := New()

	shortID, err := repo.Save("user1", "https://example.com")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if shortID == "" {
		t.Error("expected short ID, got empty string")
	}
}

func TestMemoryRepository_Save_Duplicate(t *testing.T) {
	repo := New()

	shortID1, err := repo.Save("user1", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	shortID2, err := repo.Save("user2", "https://example.com")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if shortID1 != shortID2 {
		t.Errorf("expected same short ID for duplicate URL, got %s and %s", shortID1, shortID2)
	}
}

func TestMemoryRepository_Get(t *testing.T) {
	repo := New()

	shortID, err := repo.Save("user1", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec, ok := repo.Get(shortID)
	if !ok {
		t.Error("expected to find record")
		return
	}

	if rec.OriginalURL != "https://example.com" {
		t.Errorf("expected original URL %q, got %q", "https://example.com", rec.OriginalURL)
	}

	if rec.UserID != "user1" {
		t.Errorf("expected user ID %q, got %q", "user1", rec.UserID)
	}
}

func TestMemoryRepository_Get_NotFound(t *testing.T) {
	repo := New()

	_, ok := repo.Get("nonexistent")
	if ok {
		t.Error("expected not to find record")
	}
}

func TestMemoryRepository_SaveBatch(t *testing.T) {
	repo := New()

	batch := []model.BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
	}

	responses, err := repo.SaveBatch("user1", batch)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(responses) != 2 {
		t.Errorf("expected 2 responses, got %d", len(responses))
		return
	}

	if responses[0].CorrelationID != "1" {
		t.Errorf("expected correlation ID %q, got %q", "1", responses[0].CorrelationID)
	}

	if responses[1].CorrelationID != "2" {
		t.Errorf("expected correlation ID %q, got %q", "2", responses[1].CorrelationID)
	}
}

func TestMemoryRepository_SaveBatch_Duplicate(t *testing.T) {
	repo := New()

	shortID1, _ := repo.Save("user1", "https://example.com")

	batch := []model.BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
	}

	responses, err := repo.SaveBatch("user2", batch)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if responses[0].ShortURL != shortID1 {
		t.Errorf("expected same short ID %q, got %q", shortID1, responses[0].ShortURL)
	}
}

func TestMemoryRepository_GetByUser(t *testing.T) {
	repo := New()

	repo.Save("user1", "https://example1.com")
	repo.Save("user1", "https://example2.com")
	repo.Save("user2", "https://example3.com")

	urls, err := repo.GetByUser("user1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(urls) != 2 {
		t.Errorf("expected 2 URLs, got %d", len(urls))
	}
}

func TestMemoryRepository_GetByUser_NoUser(t *testing.T) {
	repo := New()

	urls, err := repo.GetByUser("nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if urls != nil {
		t.Errorf("expected nil URLs, got %v", urls)
	}
}

func TestMemoryRepository_MarkDeleted(t *testing.T) {
	repo := New()

	shortID, _ := repo.Save("user1", "https://example.com")

	err := repo.MarkDeleted("user1", []string{shortID})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	rec, ok := repo.Get(shortID)
	if !ok {
		t.Error("expected to find record")
		return
	}

	if !rec.IsDeleted {
		t.Error("expected record to be marked as deleted")
	}
}

func TestMemoryRepository_MarkDeleted_WrongUser(t *testing.T) {
	repo := New()

	shortID, _ := repo.Save("user1", "https://example.com")

	err := repo.MarkDeleted("user2", []string{shortID})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	rec, ok := repo.Get(shortID)
	if !ok {
		t.Error("expected to find record")
		return
	}

	if rec.IsDeleted {
		t.Error("expected record not to be marked as deleted for wrong user")
	}
}

func TestMemoryRepository_MarkDeleted_Nonexistent(t *testing.T) {
	repo := New()

	err := repo.MarkDeleted("user1", []string{"nonexistent"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMemoryRepository_Ping(t *testing.T) {
	repo := New()

	err := repo.Ping()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// Benchmarks

func BenchmarkRepository_Save(b *testing.B) {
	repo := New()
	url := "https://example.com/very/long/path/with/many/segments/and/query/parameters?foo=bar&baz=qux"

	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		_, _ = repo.Save(userID, url)
	}
}

func BenchmarkRepository_Get(b *testing.B) {
	repo := New()
	shortID, _ := repo.Save(userID, "https://example.com")

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Get(shortID)
	}
}

func BenchmarkRepository_SaveBatch(b *testing.B) {
	batch := make([]model.BatchRequest, 100)
	for i := range batch {
		batch[i] = model.BatchRequest{
			CorrelationID: "corr-" + strconv.Itoa(i),
			OriginalURL:   "https://example.com/page" + strconv.Itoa(i),
		}
	}

	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		repo := New()
		_, _ = repo.SaveBatch(userID, batch)
	}
}

func BenchmarkRepository_GetByUser(b *testing.B) {
	repo := New()
	for i := 0; i < 100; i++ {
		_, _ = repo.Save(userID, "https://example.com/page"+strconv.Itoa(i))
	}

	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		_, _ = repo.GetByUser(userID)
	}
}
