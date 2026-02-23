package app

import (
	"log"
	"net/http"

	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/handler"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/logger"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/memory"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/service/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type App struct {
	cfg    *config.Config
	server *http.Server
	logger *logger.Logger
}

func New(cfg *config.Config) (*App, error) {
	logg, err := logger.NewLogger(zap.InfoLevel)
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
		return nil, err
	}

	repo := memory.New()
	r := chi.NewRouter()

	r.Use(logg.WithLogging)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	svc := url.NewURLService(repo, cfg.BaseURL)
	handler := handler.NewURLHandler(svc)
	handler.Register(r)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	return &App{
		cfg:    cfg,
		server: server,
		logger: logg,
	}, nil
}

func (a *App) Run() error {
	defer a.logger.Zap.Sync()

	a.logger.Zap.Sugar().Infow("starting server",
		"addr", a.cfg.ServerAddress,
		"url", a.cfg.BaseURL,
	)

	return a.server.ListenAndServe()
}
