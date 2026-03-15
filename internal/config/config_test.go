package config

import "testing"

func TestConfig_Defult(t *testing.T) {
	cfg := NewConfig(Config{
		ServerAddress: "",
		BaseURL:       "",
		FilePath:      "",
	})

	if cfg.ServerAddress != ":8080" {
		t.Errorf("expected default ServerAddress ':8080', got %s", cfg.ServerAddress)
	}

	if cfg.BaseURL != "http://localhost:8080/" {
		t.Errorf("expected default BaseURL 'http://localhost:8080/', got %s", cfg.BaseURL)
	}
}

func TestConfig_Flags(t *testing.T) {
	cfg := NewConfig(Config{
		ServerAddress: ":8888",
		BaseURL:       "http://flag:8888/",
		FilePath:      "",
	})

	if cfg.ServerAddress != ":8888" {
		t.Errorf("expected flag ServerAddress  ':8888', got %s", cfg.ServerAddress)
	}

	if cfg.BaseURL != "http://flag:8888/" {
		t.Errorf("expected flag BaseURL 'http://flag:8888/', got %s", cfg.BaseURL)
	}
}

func TestConfig_Env(t *testing.T) {
	t.Setenv(EnvServerAddress, ":9000")
	t.Setenv(EnvBaseURL, "http://env:9000/")

	cfg := NewConfig(Config{
		ServerAddress: "",
		BaseURL:       "",
		FilePath:      "",
	})

	if cfg.ServerAddress != ":9000" {
		t.Errorf("expected env ServerAddress ':9000', got %s", cfg.ServerAddress)
	}

	if cfg.BaseURL != "http://env:9000/" {
		t.Errorf("expected env BaseURL 'http://env:9000/', got %s", cfg.BaseURL)
	}
}
func TestConfig_EnvOverridesFlags(t *testing.T) {
	t.Setenv(EnvServerAddress, ":9000")
	t.Setenv(EnvBaseURL, "http://env:9000/")

	cfg := NewConfig(Config{
		ServerAddress: ":8888",
		BaseURL:       "http://flag:8888/",
		FilePath:      "",
	})

	if cfg.ServerAddress != ":9000" {
		t.Errorf("expected env ServerAddress ':9000', got %s", cfg.ServerAddress)
	}

	if cfg.BaseURL != "http://env:9000/" {
		t.Errorf("expected env BaseURL 'http://env:9000/', got %s", cfg.BaseURL)
	}
}
