package main

import (
	"flag"
	"log"

	"PolyakovEvg/go-musthave-shortener-tpl/internal/app"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	serverAddr := flag.String("a", ":8080", "Server address")
	baseURL := flag.String("b", "http://localhost:8080", "Base URL")
	fpath := flag.String("f", "data/storage.json", "File Path")
	dbDSN := flag.String("d", "", "DB Data Source Name")
	authSecret := flag.String("s", "", "Auth secret")

	flag.Parse()

	envErr := godotenv.Load()
	if envErr != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.NewConfig(config.Config{
		ServerAddress: *serverAddr,
		BaseURL:       *baseURL,
		FilePath:      *fpath,
		DBDSN:         *dbDSN,
		AuthSecret:    *authSecret,
	})

	a, err := app.New(cfg)

	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}

	defer a.Deleter.Close()

	if err := a.Run(); err != nil {
		log.Fatalf("app run failed: %v", err)
	}
}
