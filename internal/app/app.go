// Package app предоставляет основное приложение сервиса сокращения URL.
// Инициализирует все зависимости и запускает HTTP-сервер.
package app

import (
	"context"
	"errors"
	"net/http"

	auth "PolyakovEvg/go-musthave-shortener-tpl/internal/auth"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/handler"
	authmw "PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/auth"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/compressor"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/logger"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/db"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/file"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/memory"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/service/audit"
	service "PolyakovEvg/go-musthave-shortener-tpl/internal/service/deleter"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/service/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// App представляет основное приложение сервиса.
// Содержит конфигурацию, HTTP-сервер, логгер, хранилище и сервис удаления URL.
type App struct {
	cfg    *config.Config
	server *http.Server
	logger *logger.Logger
	repo   repository.Repository
	// Deleter — сервис асинхронного удаления URL.
	Deleter *service.Deleter
}

// New создаёт новое приложение с указанной конфигурацией.
// Инициализирует хранилище, сервисы, middleware и HTTP-обработчики.
// Возвращает ошибку, если включён HTTPS, но не указаны пути к сертификатам.
func New(cfg *config.Config) (*App, error) {
	// Валидация HTTPS
	if cfg.EnableHTTPS != nil && *cfg.EnableHTTPS {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, errors.New("HTTPS enabled but cert-file or key-file is not specified")
		}
	}

	logg, err := logger.NewLogger(zap.InfoLevel)
	if err != nil {
		return nil, err
	}

	repo, err := initRepository(cfg, logg)
	if err != nil {
		return nil, err
	}

	deleter := service.NewDeleter(repo.MarkDeleted, logg)

	authManager, err := auth.New(auth.Config{
		Secret:   cfg.AuthSecret,
		HTTPOnly: true,
		Secure:   false,
	})
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()

	r.Use(authmw.WithCookie(authManager))
	r.Use(compressor.WithGzip)
	r.Use(logg.WithLogging)
	r.Use(middleware.Recoverer)

	urlService := url.NewURLService(repo, cfg.BaseURL)
	auditService := audit.NewAuditService(logg)
	initObservers(cfg, logg, auditService)

	urlHandler := handler.NewURLHandler(urlService, auditService, deleter, cfg, logg)
	urlHandler.Register(r)

	statsHandler := handler.NewStatsHandler(repo, cfg.TrustedSubnet)
	r.Get("/api/internal/stats", statsHandler.GetStats)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	return &App{
		cfg:     cfg,
		server:  server,
		logger:  logg,
		repo:    repo,
		Deleter: deleter,
	}, nil
}

// Run запускает HTTP-сервер.
// Если включён HTTPS (EnableHTTPS), использует ListenAndServeTLS.
func (a *App) Run() error {
	a.logger.Zap.Infow("starting server",
		"addr", a.cfg.ServerAddress,
		"url", a.cfg.BaseURL,
		"https", a.cfg.EnableHTTPS != nil && *a.cfg.EnableHTTPS,
	)

	if a.cfg.EnableHTTPS != nil && *a.cfg.EnableHTTPS {
		return a.server.ListenAndServeTLS(a.cfg.CertFile, a.cfg.KeyFile)
	}
	return a.server.ListenAndServe()
}

// Shutdown останавливает HTTP-сервер и закрывает все ресурсы.
// Принимает контекст для таймаута завершения.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Zap.Info("shutting down server gracefully...")

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Zap.Errorf("server shutdown error: %v", err)
	}

	a.Deleter.Close()

	if err := a.repo.Close(); err != nil {
		a.logger.Zap.Errorf("repository close error: %v", err)
	}

	a.logger.Zap.Sync()
	a.logger.Zap.Info("server stopped")

	return nil
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

func initObservers(cfg *config.Config, logg *logger.Logger, auditService *audit.AuditService) {
	fileObserver, err := audit.NewFileObserver(cfg.AuditFile)
	if err != nil {
		logg.Zap.Infow("File observer error", "error", err)
	} else {
		auditService.Register(fileObserver)
	}

	httpObserver, err := audit.NewHTTPObserver(cfg.AuditURL)
	if err != nil {
		logg.Zap.Infow("HTTP observer error", "error", err)
	} else {
		auditService.Register(httpObserver)
	}
}
