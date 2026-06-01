// Package config предоставляет конфигурацию для сервиса сокращения URL.
// Поддерживает загрузку параметров из переменных окружения и флагов командной строки.
package config

import (
	"os"
	"strings"
)

// Config содержит параметры конфигурации сервиса.
type Config struct {
	// ServerAddress — адрес сервера в формате host:port (например, ":8080").
	ServerAddress string
	// BaseURL — базовый URL для сокращённых ссылок (например, "http://localhost:8080/").
	BaseURL string
	// FilePath — путь к файлу для хранения данных (для файлового хранилища).
	FilePath string
	// DBDSN — строка подключения к базе данных (для DB хранилища).
	DBDSN string
	// AuthSecret — секретный ключ для подписи JWT токенов.
	AuthSecret string
	// AuditFile — путь к файлу для записи событий аудита.
	AuditFile string
	// AuditURL — URL для отправки событий аудита по HTTP.
	AuditURL string
}

// Константы для имён переменных окружения.
const (
	// EnvServerAddress — переменная окружения для адреса сервера.
	EnvServerAddress string = "SERVER_ADDRESS"
	// EnvBaseURL — переменная окружения для базового URL.
	EnvBaseURL string = "BASE_URL"
	// EnvFilePath — переменная окружения для пути к файлу хранилища.
	EnvFilePath string = "FILE_STORAGE_PATH"
	// EnvDB — переменная окружения для строки подключения к БД.
	EnvDB string = "DATABASE_DSN"
	// EnvAuthSecret — переменная окружения для секретного ключа.
	EnvAuthSecret string = "AUTH_SECRET"
	// EnvAuditFile — переменная окружения для файла аудита.
	EnvAuditFile string = "AUDIT_FILE"
	// EnvAuditURL — переменная окружения для URL аудита.
	EnvAuditURL string = "AUDIT_URL"
)

// NewConfig создаёт новую конфигурацию с значениями по умолчанию.
// Приоритет: переменные окружения > переданные параметры > значения по умолчанию.
func NewConfig(params Config) *Config {
	cfg := &Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080/",
		FilePath:      "data/storage.json",
		DBDSN:         "",
		AuthSecret:    "",
	}

	envAddr, ok := os.LookupEnv(EnvServerAddress)
	if ok {
		cfg.ServerAddress = envAddr
	} else if params.ServerAddress != "" {
		cfg.ServerAddress = params.ServerAddress
	}

	envBaseURL, ok := os.LookupEnv(EnvBaseURL)
	if ok {
		cfg.BaseURL = envBaseURL
	} else if params.BaseURL != "" {
		cfg.BaseURL = params.BaseURL
	}

	envFilePath, ok := os.LookupEnv(EnvFilePath)
	if ok {
		cfg.FilePath = envFilePath
	} else if params.FilePath != "" {
		cfg.FilePath = params.FilePath
	}

	envDB, ok := os.LookupEnv(EnvDB)
	if ok {
		cfg.DBDSN = envDB
	} else if params.DBDSN != "" {
		cfg.DBDSN = params.DBDSN
	}

	envAuthSecret, ok := os.LookupEnv(EnvAuthSecret)
	if ok {
		cfg.AuthSecret = envAuthSecret
	} else if params.AuthSecret != "" {
		cfg.AuthSecret = params.AuthSecret
	}

	envAuditFile, ok := os.LookupEnv(EnvAuditFile)
	if ok {
		cfg.AuditFile = envAuditFile
	} else if params.AuditFile != "" {
		cfg.AuditFile = params.AuditFile
	}

	envAuditURL, ok := os.LookupEnv(EnvAuditURL)
	if ok {
		cfg.AuditURL = envAuditURL
	} else if params.AuditURL != "" {
		cfg.AuditURL = params.AuditURL
	}

	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/") + "/"

	return cfg
}
