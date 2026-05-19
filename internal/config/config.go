package config

import (
	"os"
	"strings"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	FilePath      string
	DBDSN         string
	AuthSecret    string
	AuditFile     string
	AuditURL      string
}

const (
	EnvServerAddress string = "SERVER_ADDRESS"
	EnvBaseURL       string = "BASE_URL"
	EnvFilePath      string = "FILE_STORAGE_PATH"
	EnvDB            string = "DATABASE_DSN"
	EnvAuthSecret    string = "AUTH_SECRET"
	EnvAuditFile     string = "AUDIT_FILE"
	EnvAuditURL      string = "AUDIT_URL"
)

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
