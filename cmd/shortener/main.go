package main

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/handler"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/memory"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	serverAddr := flag.String("a", "", "Input server ServerAddress")
	baseURL := flag.String("b", "", "Input server BaseURL")
	flag.Parse()

	fmt.Println(*serverAddr, *baseURL)

	cfg := config.NewConfig(*serverAddr, *baseURL)
	repo := memory.New()
	mux := handler.Router(repo, cfg)

	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, mux))
}
