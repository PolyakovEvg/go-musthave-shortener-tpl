// Package logger предоставляет middleware для логирования HTTP-запросов.
// Логирует URI, метод, статус, длительность и размер ответа.
package logger

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger — обёртка над zap.SugaredLogger для логирования запросов.
type Logger struct {
	// Zap — SugaredLogger для структурированного логирования.
	Zap *zap.SugaredLogger
}

// ResponseData содержит данные ответа для логирования.
type (
	ResponseData struct {
		// Status — HTTP-статус ответа.
		Status int
		// Size — размер ответа в байтах.
		Size int
	}

	// LoggingResponseWriter — обёртка над http.ResponseWriter для перехвата данных ответа.
	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

// NewLogger создаёт новый логгер с указанным уровнем логирования.
func NewLogger(lvl zapcore.Level) (*Logger, error) {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(lvl)

	cfg := zap.NewProductionConfig()
	cfg.Level = atom

	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed cfg.Build: %v", err)
	}

	return &Logger{Zap: zl.Sugar()}, nil
}

// Write записывает данные в ответ и обновляет счётчик размера.
func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size
	return size, err
}

// WriteHeader записывает статус-код ответа и сохраняет его для логирования.
func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode
}

// WithLogging возвращает middleware для логирования HTTP-запросов и ответов.
// Логирует URI, метод, статус, длительность и размер ответа.
func (logger *Logger) WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &ResponseData{
			Status: 200,
			Size:   0,
		}
		lw := LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Zap.Infow(
			"request",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.Status,
			"duration", duration,
			"size", responseData.Size,
		)
	})
}
