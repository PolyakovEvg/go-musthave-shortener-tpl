package config

import (
	"os"
	"strings"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

const (
	EnvServerAddress string = "SERVER_ADDRESS"
	EnvBaseURL       string = "BASE_URL"
)

func NewConfig(flagAddr, flagBaseURL string) *Config {
	cfg := &Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080/",
	}

	envAddr := os.Getenv(EnvServerAddress)
	envBaseURL := os.Getenv(EnvBaseURL)

	if envAddr != "" {
		cfg.ServerAddress = envAddr
	} else if flagAddr != "" {
		cfg.ServerAddress = flagAddr
	}

	if envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	} else if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}

	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/") + "/"

	return cfg
}
