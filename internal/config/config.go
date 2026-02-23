package config

import "os"

type Config struct {
	ServerAddress string
	BaseURL       string
}

const (
	EnvServerAddress = "SERVER_ADDRESS"
	EnvBaseURL       = "BASE_URL"
)

func NewConfig(flagAddr, flagBaseURL string) *Config {
	cfg := &Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080/",
	}

	envAddr := os.Getenv(EnvServerAddress)
	envBase := os.Getenv(EnvBaseURL)

	if envAddr != "" {
		cfg.ServerAddress = envAddr
	} else if flagAddr != "" {
		cfg.ServerAddress = flagAddr
	}

	if envBase != "" {
		cfg.BaseURL = envBase
	} else if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}

	return cfg
}
