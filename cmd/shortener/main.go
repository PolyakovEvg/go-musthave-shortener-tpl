package main

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/handler"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"log"
	"net/http"
)

func main() {
	cfg := config.NewConfig()
	storage := repository.NewStorage()

	mux := handler.Router(storage, cfg)

	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, mux))
}
