package config

import (
	"os"
	"strings"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	FilePath      string
	AuthSecret    string
}

const (
	EnvServerAddress string = "SERVER_ADDRESS"
	EnvBaseURL       string = "BASE_URL"
	EnvFilePath      string = "FILE_STORAGE_PATH"
	EnvAuthSecret    string = "AUTH_SECRET"
)

func NewConfig(flagAddr, flagBaseURL, flagFilePath, flatAuthSecret string) *Config {
	cfg := &Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080/",
		FilePath:      "data/storage.json",
		AuthSecret:    "",
	}

	envAddr, ok := os.LookupEnv(EnvServerAddress)
	if ok {
		cfg.ServerAddress = envAddr
	} else if flagAddr != "" {
		cfg.ServerAddress = flagAddr
	}

	envBaseURL, ok := os.LookupEnv(EnvBaseURL)
	if ok {
		cfg.BaseURL = envBaseURL
	} else if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}

	envFilePath, ok := os.LookupEnv(EnvFilePath)
	if ok {
		cfg.FilePath = envFilePath
	} else if flagFilePath != "" {
		cfg.FilePath = flagFilePath
	}

	envAuthSecret, ok := os.LookupEnv(EnvAuthSecret)
	if ok {
		cfg.AuthSecret = envAuthSecret
	} else if flatAuthSecret != "" {
		cfg.AuthSecret = flatAuthSecret
	}

	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/") + "/"

	return cfg
}
