package main

import (
	"flag"
	"log"

	"PolyakovEvg/go-musthave-shortener-tpl/internal/app"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
)

func main() {
	serverAddr := flag.String("a", ":8080", "Server address")
	baseURL := flag.String("b", "http://localhost:8080", "Base URL")
	fpath := flag.String("f", "data/storage.json", "File Path")
	authSecret := flag.String("s", "test_auth_secret", "Auth secret")

	flag.Parse()

	cfg := config.NewConfig(*serverAddr, *baseURL, *fpath, *authSecret)

	a, err := app.New(cfg)

	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}

	if err := a.Run(); err != nil {
		log.Fatalf("app run failed: %v", err)
	}
}
