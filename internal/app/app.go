// Package app предоставляет основное приложение сервиса сокращения URL.
// Инициализирует все зависимости и запускает HTTP-сервер.
package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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
func New(cfg *config.Config) (*App, error) {
	logg, err := logger.NewLogger(zap.InfoLevel)
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
		return nil, err
	}

	repo, err := initRepository(cfg, logg)
	if err != nil {
		logg.Zap.Fatalf("can't initialize repository: %v", err)
	}

	deleter := service.NewDeleter(repo.MarkDeleted, logg)

	authManager, err := auth.New(auth.Config{
		Secret:   cfg.AuthSecret,
		HTTPOnly: true,
		Secure:   false,
	})
	if err != nil {
		logg.Zap.Fatalf("can't initialize auth manager %v", err)
	}

	r := chi.NewRouter()

	r.Use(authmw.WithCookie(authManager))
	r.Use(compressor.WithGzip)
	r.Use(logg.WithLogging)
	r.Use(middleware.Recoverer)

	urlService := url.NewURLService(repo, cfg.BaseURL)
	auditService := audit.NewAuditService(logg)
	initObservers(cfg, logg, auditService)

	handler := handler.NewURLHandler(urlService, auditService, deleter, cfg, logg)
	handler.Register(r)

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

// Run запускает HTTP-сервер и блокирует выполнение до остановки сервера.
// Обрабатывает сигналы SIGTERM, SIGINT, SIGQUIT для graceful shutdown.
// При завершении сохраняет несохранённые данные и синхронизирует логи.
// Если включён HTTPS (EnableHTTPS), использует ListenAndServeTLS.
func (a *App) Run() error {
	defer a.logger.Zap.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	go func() {
		a.logger.Zap.Infow("starting server",
			"addr", a.cfg.ServerAddress,
			"url", a.cfg.BaseURL,
			"https", a.cfg.EnableHTTPS,
		)

		var err error
		if a.cfg.EnableHTTPS {
			err = a.server.ListenAndServeTLS(a.cfg.CertFile, a.cfg.KeyFile)
		} else {
			err = a.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			a.logger.Zap.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()

	a.logger.Zap.Info("shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Zap.Errorf("server shutdown error: %v", err)
	}

	a.Deleter.Close()

	if err := a.repo.Close(); err != nil {
		a.logger.Zap.Errorf("repository close error: %v", err)
	}

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
