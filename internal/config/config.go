// Package config предоставляет конфигурацию для сервиса сокращения URL.
// Поддерживает загрузку параметров из JSON-файла, переменных окружения и флагов командной строки.
package config

import (
	"encoding/json"
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
	// EnableHTTPS — флаг включения HTTPS.
	EnableHTTPS bool
	// CertFile — путь к файлу сертификата TLS.
	CertFile string
	// KeyFile — путь к файлу приватного ключа TLS.
	KeyFile string
	// ConfigFile — путь к файлу конфигурации JSON.
	ConfigFile string
}

// jsonConfig представляет структуру JSON-файла конфигурации.
type jsonConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
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
	// EnvEnableHTTPS — переменная окружения для включения HTTPS.
	EnvEnableHTTPS string = "ENABLE_HTTPS"
	// EnvCertFile — переменная окружения для пути к файлу сертификата.
	EnvCertFile string = "CERT_FILE"
	// EnvKeyFile — переменная окружения для пути к файлу ключа.
	EnvKeyFile string = "KEY_FILE"
	// EnvConfig — переменная окружения для пути к файлу конфигурации.
	EnvConfig string = "CONFIG"
)

// NewConfig создаёт новую конфигурацию с значениями по умолчанию.
// Приоритет: переменные окружения > флаги > JSON-файл > значения по умолчанию.
func NewConfig(params Config) *Config {
	cfg := &Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080/",
		FilePath:      "data/storage.json",
		DBDSN:         "",
		AuthSecret:    "",
	}

	configFile, ok := os.LookupEnv(EnvConfig)
	if !ok && params.ConfigFile != "" {
		configFile = params.ConfigFile
	}

	if configFile != "" {
		loadJSONConfig(cfg, configFile)
	}

	applyParams(cfg, params)

	applyEnvVars(cfg)

	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/") + "/"

	return cfg
}

func loadJSONConfig(cfg *Config, filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var jc jsonConfig
	if err := json.Unmarshal(data, &jc); err != nil {
		return
	}

	if jc.ServerAddress != "" {
		cfg.ServerAddress = jc.ServerAddress
	}
	if jc.BaseURL != "" {
		cfg.BaseURL = jc.BaseURL
	}
	if jc.FileStoragePath != "" {
		cfg.FilePath = jc.FileStoragePath
	}
	if jc.DatabaseDSN != "" {
		cfg.DBDSN = jc.DatabaseDSN
	}
	cfg.EnableHTTPS = jc.EnableHTTPS
}

func applyParams(cfg *Config, params Config) {
	if params.ServerAddress != "" {
		cfg.ServerAddress = params.ServerAddress
	}
	if params.BaseURL != "" {
		cfg.BaseURL = params.BaseURL
	}
	if params.FilePath != "" {
		cfg.FilePath = params.FilePath
	}
	if params.DBDSN != "" {
		cfg.DBDSN = params.DBDSN
	}
	if params.AuthSecret != "" {
		cfg.AuthSecret = params.AuthSecret
	}
	if params.AuditFile != "" {
		cfg.AuditFile = params.AuditFile
	}
	if params.AuditURL != "" {
		cfg.AuditURL = params.AuditURL
	}
	if params.EnableHTTPS {
		cfg.EnableHTTPS = true
	}
	if params.CertFile != "" {
		cfg.CertFile = params.CertFile
	}
	if params.KeyFile != "" {
		cfg.KeyFile = params.KeyFile
	}
}

func applyEnvVars(cfg *Config) {
	if envAddr, ok := os.LookupEnv(EnvServerAddress); ok {
		cfg.ServerAddress = envAddr
	}
	if envBaseURL, ok := os.LookupEnv(EnvBaseURL); ok {
		cfg.BaseURL = envBaseURL
	}
	if envFilePath, ok := os.LookupEnv(EnvFilePath); ok {
		cfg.FilePath = envFilePath
	}
	if envDB, ok := os.LookupEnv(EnvDB); ok {
		cfg.DBDSN = envDB
	}
	if envAuthSecret, ok := os.LookupEnv(EnvAuthSecret); ok {
		cfg.AuthSecret = envAuthSecret
	}
	if envAuditFile, ok := os.LookupEnv(EnvAuditFile); ok {
		cfg.AuditFile = envAuditFile
	}
	if envAuditURL, ok := os.LookupEnv(EnvAuditURL); ok {
		cfg.AuditURL = envAuditURL
	}
	if envEnableHTTPS, ok := os.LookupEnv(EnvEnableHTTPS); ok && envEnableHTTPS == "true" {
		cfg.EnableHTTPS = true
	}
	if envCertFile, ok := os.LookupEnv(EnvCertFile); ok {
		cfg.CertFile = envCertFile
	}
	if envKeyFile, ok := os.LookupEnv(EnvKeyFile); ok {
		cfg.KeyFile = envKeyFile
	}
}
