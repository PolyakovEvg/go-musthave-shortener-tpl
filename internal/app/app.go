package app

import (
	"log"
	"net/http"

	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/handler"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/compressor"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/logger"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/db"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/file"
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

	repo, err := initRepository(cfg, logg)
	if err != nil {
		logg.Zap.Fatalf("can't initialize file repository %v", err)
	}

	r := chi.NewRouter()

	r.Use(compressor.WithGzip)
	r.Use(logg.WithLogging)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	svc := url.NewURLService(repo, cfg.BaseURL)
	handler := handler.NewURLHandler(svc, cfg, logg)
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

	a.logger.Zap.Infow("starting server",
		"addr", a.cfg.ServerAddress,
		"url", a.cfg.BaseURL,
	)

	return a.server.ListenAndServe()
}

func initRepository(cfg *config.Config, logg *logger.Logger) (repository.Repository, error) {
	if cfg.DBDSN != "" {
		logg.Zap.Info("using database storage")
		return db.New(cfg.DBDSN)
	}

	if cfg.FilePath != "" {
		logg.Zap.Info("using file storage")
		return file.New(cfg.FilePath)
	}

	logg.Zap.Info("using in-memory storage")
	return memory.New(), nil
}
