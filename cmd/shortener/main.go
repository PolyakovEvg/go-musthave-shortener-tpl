// Команда shortener запускает сервис сокращения URL.
//
// Сервис предоставляет HTTP API для создания коротких URL и перенаправления по ним.
// Поддерживает несколько типов хранилищ: in-memory, файл и базу данных.
//
// Флаги командной строки:
//
//	-a — адрес сервера (по умолчанию ":8080")
//	-b — базовый URL для коротких ссылок (по умолчанию "http://localhost:8080")
//	-f — путь к файлу хранилища (по умолчанию "data/storage.json")
//	-d — строка подключения к БД (по умолчанию пусто)
//	-s — включить HTTPS (требует -cert-file и -key-file)
//	-secret — секретный ключ для JWT (по умолчанию пусто)
//	-audit-file — путь к файлу аудита
//	-audit-url — URL для отправки событий аудита
//	-cert-file — путь к файлу сертификата TLS
//	-key-file — путь к файлу приватного ключа TLS
//	-c, -config — путь к файлу конфигурации JSON
//	-t — доверенная подсеть в формате CIDR (например, "192.168.1.0/24")
//
// Приоритет конфигурации: переменные окружения > флаги > JSON-файл > значения по умолчанию.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"PolyakovEvg/go-musthave-shortener-tpl/internal/app"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()
	serverAddr := flag.String("a", "", "Server address")
	baseURL := flag.String("b", "", "Base URL")
	fpath := flag.String("f", "", "File Path")
	dbDSN := flag.String("d", "", "DB Data Source Name")
	enableHTTPS := flag.Bool("s", false, "Enable HTTPS")
	authSecret := flag.String("secret", "", "Auth secret")
	auditFile := flag.String("audit-file", "", "Audit file path")
	auditURL := flag.String("audit-url", "", "Audit remote URL")
	certFile := flag.String("cert-file", "", "Path to TLS certificate file")
	keyFile := flag.String("key-file", "", "Path to TLS private key file")
	configFile := flag.String("c", "", "Path to JSON config file")
	flag.StringVar(configFile, "config", "", "Path to JSON config file (alias for -c)")
	trustedSubnet := flag.String("t", "", "Trusted subnet in CIDR format")

	flag.Parse()

	envErr := godotenv.Load()
	if envErr != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.NewConfig(config.Config{
		ServerAddress: *serverAddr,
		BaseURL:       *baseURL,
		FilePath:      *fpath,
		DBDSN:         *dbDSN,
		AuthSecret:    *authSecret,
		AuditFile:     *auditFile,
		AuditURL:      *auditURL,
		EnableHTTPS:   getFlagIfSet(enableHTTPS, "s"),
		CertFile:      *certFile,
		KeyFile:       *keyFile,
		ConfigFile:    *configFile,
		TrustedSubnet: *trustedSubnet,
	})

	if err != nil {
		log.Fatalf("config init failed: %v", err)
	}

	a, err := app.New(cfg)

	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := a.Run(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		log.Println("received shutdown signal")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return a.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func printBuildInfo() {
	log.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		getBuildValue(buildVersion),
		getBuildValue(buildDate),
		getBuildValue(buildCommit))
}

func getBuildValue(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}

// getFlagIfSet возвращает указатель на значение флага, если он был явно задан.
// Используется для различения "флаг не задан" (nil) и "флаг задан как false".
func getFlagIfSet[T any](flagValue *T, flagName string) *T {
	var result *T
	flag.Visit(func(f *flag.Flag) {
		if f.Name == flagName {
			result = flagValue
		}
	})
	return result
}
