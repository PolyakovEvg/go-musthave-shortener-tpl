package main

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/handler"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	serverAddr := flag.String("a", "", "Input server ServerAddress")
	baseUrl := flag.String("b", "", "Input server BaseURL")
	flag.Parse()

	fmt.Println(*serverAddr, *baseUrl)

	cfg := config.NewConfig(*serverAddr, *baseUrl)
	storage := repository.NewStorage()
	mux := handler.Router(storage, cfg)

	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, mux))
}
