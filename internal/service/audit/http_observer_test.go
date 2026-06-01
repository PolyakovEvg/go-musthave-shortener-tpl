package audit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPObserver(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errType error
	}{
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errType: ErrEmptyURL,
		},
		{
			name:    "valid URL",
			url:     "http://localhost:8080/audit",
			wantErr: false,
		},
		{
			name:    "HTTPS URL",
			url:     "https://example.com/api/audit",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observer, err := NewHTTPObserver(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewHTTPObserver() expected error, got nil")
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewHTTPObserver() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("NewHTTPObserver() unexpected error: %v", err)
			}

			if observer == nil {
				t.Errorf("NewHTTPObserver() returned nil observer")
			}

			if observer.url != tt.url {
				t.Errorf("observer.url = %v, want %v", observer.url, tt.url)
			}

			if observer.client == nil {
				t.Errorf("observer.client is nil")
			}

			if observer.client.Timeout != 5*time.Second {
				t.Errorf("client.Timeout = %v, want 5s", observer.client.Timeout)
			}
		})
	}
}

func TestHTTPObserver_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var event AuditEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer, err := NewHTTPObserver(server.URL)
	if err != nil {
		t.Fatalf("Failed to create observer: %v", err)
	}

	event := AuditEvent{
		Ts:     time.Now().Unix(),
		Action: "shorten",
		UserID: "user123",
		URL:    "https://example.com",
	}

	err = observer.Send(event)
	if err != nil {
		t.Errorf("Send() unexpected error: %v", err)
	}
}

func TestHTTPObserver_Send_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	observer, err := NewHTTPObserver(server.URL)
	if err != nil {
		t.Fatalf("Failed to create observer: %v", err)
	}

	event := AuditEvent{
		Ts:     time.Now().Unix(),
		Action: "shorten",
		URL:    "https://example.com",
	}

	err = observer.Send(event)
	if err != nil {
		t.Errorf("Send() unexpected error: %v", err)
	}
}

func TestHTTPObserver_Send_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer, err := NewHTTPObserver(server.URL)
	if err != nil {
		t.Fatalf("Failed to create observer: %v", err)
	}

	observer.client.Timeout = 500 * time.Millisecond

	event := AuditEvent{
		Ts:     time.Now().Unix(),
		Action: "shorten",
		URL:    "https://example.com",
	}

	err = observer.Send(event)
	if err == nil {
		t.Errorf("Send() expected timeout error, got nil")
	}
}

func TestHTTPObserver_Send_InvalidURL(t *testing.T) {
	observer, err := NewHTTPObserver("http://localhost:99999/invalid")
	if err != nil {
		t.Fatalf("Failed to create observer: %v", err)
	}

	event := AuditEvent{
		Ts:     time.Now().Unix(),
		Action: "shorten",
		URL:    "https://example.com",
	}

	err = observer.Send(event)
	if err == nil {
		t.Errorf("Send() expected error for invalid URL, got nil")
	}
}

func TestHTTPObserver_Send_UnreachableHost(t *testing.T) {
	observer, err := NewHTTPObserver("http://unreachable.host.local:9999/audit")
	if err != nil {
		t.Fatalf("Failed to create observer: %v", err)
	}

	event := AuditEvent{
		Ts:     time.Now().Unix(),
		Action: "shorten",
		URL:    "https://example.com",
	}

	err = observer.Send(event)
	if err == nil {
		t.Errorf("Send() expected error for unreachable host, got nil")
	}
}

func TestHTTPObserver_Send_RequestData(t *testing.T) {
	receivedEvent := &AuditEvent{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(receivedEvent); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer, err := NewHTTPObserver(server.URL)
	if err != nil {
		t.Fatalf("Failed to create observer: %v", err)
	}

	expectedEvent := AuditEvent{
		Ts:     1234567890,
		Action: "follow",
		UserID: "test-user",
		URL:    "https://test.com/path?query=1",
	}

	err = observer.Send(expectedEvent)
	if err != nil {
		t.Errorf("Send() error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if receivedEvent.Ts != expectedEvent.Ts {
		t.Errorf("Ts = %d, want %d", receivedEvent.Ts, expectedEvent.Ts)
	}
	if receivedEvent.Action != expectedEvent.Action {
		t.Errorf("Action = %s, want %s", receivedEvent.Action, expectedEvent.Action)
	}
	if receivedEvent.UserID != expectedEvent.UserID {
		t.Errorf("UserID = %s, want %s", receivedEvent.UserID, expectedEvent.UserID)
	}
	if receivedEvent.URL != expectedEvent.URL {
		t.Errorf("URL = %s, want %s", receivedEvent.URL, expectedEvent.URL)
	}
}
