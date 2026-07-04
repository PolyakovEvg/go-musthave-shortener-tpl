// Package app предоставляет основное приложение сервиса сокращения URL.
// Инициализирует все зависимости и запускает HTTP- и gRPC-серверы.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	auth "PolyakovEvg/go-musthave-shortener-tpl/internal/auth"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"
	grpcserver "PolyakovEvg/go-musthave-shortener-tpl/internal/grpc"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// App представляет основное приложение сервиса.
// Содержит конфигурацию, HTTP-сервер, gRPC-сервер, логгер, хранилище и сервис удаления URL.
type App struct {
	cfg        *config.Config
	server     *http.Server
	grpcServer *grpc.Server
	grpcAddr   string
	logger     *logger.Logger
	repo       repository.Repository
	urlService *url.URLService
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

	var grpcSrv *grpc.Server
	if cfg.EnableHTTPS != nil && *cfg.EnableHTTPS && cfg.CertFile != "" && cfg.KeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials for gRPC: %w", err)
		}
		grpcSrv = grpc.NewServer(
			grpc.Creds(creds),
			grpc.UnaryInterceptor(grpcserver.AuthInterceptor(authManager)),
		)
	} else {
		grpcSrv = grpc.NewServer(
			grpc.UnaryInterceptor(grpcserver.AuthInterceptor(authManager)),
		)
	}
	grpcserver.Register(grpcSrv, urlService, authManager)

	return &App{
		cfg:        cfg,
		server:     server,
		grpcServer: grpcSrv,
		grpcAddr:   cfg.GRPCAddress,
		logger:     logg,
		repo:       repo,
		urlService: urlService,
		Deleter:    deleter,
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

// RunGRPC запускает gRPC-сервер.
func (a *App) RunGRPC() error {
	if a.grpcAddr == "" {
		return nil
	}

	a.logger.Zap.Infow("starting gRPC server", "addr", a.grpcAddr)

	listener, err := net.Listen("tcp", a.grpcAddr)
	if err != nil {
		return err
	}

	return a.grpcServer.Serve(listener)
}

// Shutdown останавливает HTTP-сервер и закрывает все ресурсы.
// Принимает контекст для таймаута завершения.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Zap.Info("shutting down server gracefully...")

	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

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
