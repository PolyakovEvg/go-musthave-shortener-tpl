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
	flag.Parse()

	cfg := config.NewConfig(*serverAddr, *baseURL)

	a, err := app.New(cfg)

	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}

	if err := a.Run(); err != nil {
		log.Fatalf("app run failed: %v", err)
	}
}
