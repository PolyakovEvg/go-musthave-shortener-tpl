package config

type Config struct {
	ServerAddress string
	BaseURL       string
}

const urlSuffix string = "/"

func NewConfig(serverAddr, baseURL string) *Config {
	cfg := &Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080/",
	}

	if serverAddr != "" {
		cfg.ServerAddress = serverAddr
	}

	if baseURL != "" {
		cfg.BaseURL = baseURL + urlSuffix
	}

	return cfg
}
