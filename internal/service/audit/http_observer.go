package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var ErrEmptyURL = errors.New("audit URL is empty")

type HTTPObserver struct {
	url    string
	client *http.Client
}

func NewHTTPObserver(url string) (*HTTPObserver, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	return &HTTPObserver{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (o *HTTPObserver) Send(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
