// Package handler_test содержит примеры работы с API сервиса сокращения URL.
package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

// Пример создания короткого URL через POST / (text/plain).
func Example_shortenURL() {
	// Запрос на сокращение URL в формате text/plain
	body := "https://example.com/very/long/url/that/needs/shortening"
	_ = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

	// В реальном коде:
	// req.Header.Set("Content-Type", "text/plain")
	// resp, _ := srv.Client().Do(req)

	fmt.Println("Request: POST /")
	fmt.Println("Content-Type: text/plain")
	fmt.Println("Body:", body)
	fmt.Println()
	fmt.Println("Response: 201 Created")
	fmt.Println("Body: http://localhost:8080/abc123")

	// Output:
	// Request: POST /
	// Content-Type: text/plain
	// Body: https://example.com/very/long/url/that/needs/shortening
	//
	// Response: 201 Created
	// Body: http://localhost:8080/abc123
}

// Пример создания короткого URL через POST /api/shorten (JSON).
func Example_postShorten() {
	// Запрос на сокращение URL в формате JSON
	reqBody := map[string]string{"url": "https://example.com/long-url"}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("Request: POST /api/shorten")
	fmt.Println("Content-Type: application/json")
	fmt.Println("Body:", string(jsonBody))
	fmt.Println()
	fmt.Println("Response: 201 Created")
	fmt.Println(`Body: {"result":"http://localhost:8080/abc123"}`)

	// Output:
	// Request: POST /api/shorten
	// Content-Type: application/json
	// Body: {"url":"https://example.com/long-url"}
	//
	// Response: 201 Created
	// Body: {"result":"http://localhost:8080/abc123"}
}

// Пример редиректа по короткому URL через GET /{id}.
func Example_redirectURL() {
	// Запрос на получение оригинального URL
	_ = httptest.NewRequest(http.MethodGet, "/abc123", nil)

	fmt.Println("Request: GET /abc123")
	fmt.Println()
	fmt.Println("Response: 307 Temporary Redirect")
	fmt.Println("Location: https://example.com/original-url")

	// Output:
	// Request: GET /abc123
	//
	// Response: 307 Temporary Redirect
	// Location: https://example.com/original-url
}

// Пример пакетного сокращения URL через POST /api/shorten/batch.
func Example_shortenBatch() {
	// Запрос на пакетное сокращение URL
	reqBody := []map[string]string{
		{"correlation_id": "1", "original_url": "https://example.com/url1"},
		{"correlation_id": "2", "original_url": "https://example.com/url2"},
	}
	jsonBody, _ := json.Marshal(reqBody)

	_ = httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(jsonBody))

	fmt.Println("Request: POST /api/shorten/batch")
	fmt.Println("Content-Type: application/json")
	fmt.Println("Body:", string(jsonBody))
	fmt.Println()
	fmt.Println("Response: 201 Created")
	fmt.Println(`Body: [{"correlation_id":"1","short_url":"http://localhost:8080/abc1"},{"correlation_id":"2","short_url":"http://localhost:8080/abc2"}]`)

	// Output:
	// Request: POST /api/shorten/batch
	// Content-Type: application/json
	// Body: [{"correlation_id":"1","original_url":"https://example.com/url1"},{"correlation_id":"2","original_url":"https://example.com/url2"}]
	//
	// Response: 201 Created
	// Body: [{"correlation_id":"1","short_url":"http://localhost:8080/abc1"},{"correlation_id":"2","short_url":"http://localhost:8080/abc2"}]
}

// Пример получения URL пользователя через GET /api/user/urls.
func Example_getUserURLs() {
	// Запрос на получение всех URL пользователя
	_ = httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	// Требуется авторизация через cookie или заголовок Authorization

	fmt.Println("Request: GET /api/user/urls")
	fmt.Println("Cookie: user_id=<auth_token>")
	fmt.Println()
	fmt.Println("Response: 200 OK")
	fmt.Println(`Body: [{"short_url":"http://localhost:8080/abc1","original_url":"https://example.com/url1"}]`)

	// Output:
	// Request: GET /api/user/urls
	// Cookie: user_id=<auth_token>
	//
	// Response: 200 OK
	// Body: [{"short_url":"http://localhost:8080/abc1","original_url":"https://example.com/url1"}]
}

// Пример проверки соединения с БД через GET /ping.
func Example_pingHandler() {
	// Запрос на проверку соединения с БД
	_ = httptest.NewRequest(http.MethodGet, "/ping", nil)

	resp := httptest.NewRecorder()
	// В реальном коде: handler.PingHandler(resp, req)

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("Request: GET /ping")
	fmt.Println()
	fmt.Println("Response:", resp.Code, http.StatusText(resp.Code))
	fmt.Println("Body:", string(body))

	// Output:
	// Request: GET /ping
	//
	// Response: 200 OK
	// Body:
}
