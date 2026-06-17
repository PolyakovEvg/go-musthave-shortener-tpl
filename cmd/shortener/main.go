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
//	-s — включить HTTPS (по умолчанию false)
//	-secret — секретный ключ для JWT (по умолчанию пусто)
//	-audit-file — путь к файлу аудита
//	-audit-url — URL для отправки событий аудита
//	-cert-file — путь к файлу сертификата TLS
//	-key-file — путь к файлу приватного ключа TLS
//	-c, -config — путь к файлу конфигурации JSON
//
// Приоритет конфигурации: переменные окружения > флаги > JSON-файл > значения по умолчанию.
package main

import (
	"flag"
	"log"

	"PolyakovEvg/go-musthave-shortener-tpl/internal/app"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/config"

	"github.com/joho/godotenv"
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

	flag.Parse()

	envErr := godotenv.Load()
	if envErr != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.NewConfig(config.Config{
		ServerAddress: *serverAddr,
		BaseURL:       *baseURL,
		FilePath:      *fpath,
		DBDSN:         *dbDSN,
		AuthSecret:    *authSecret,
		AuditFile:     *auditFile,
		AuditURL:      *auditURL,
		EnableHTTPS:   *enableHTTPS,
		CertFile:      *certFile,
		KeyFile:       *keyFile,
		ConfigFile:    *configFile,
	})

	a, err := app.New(cfg)

	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}

	defer a.Deleter.Close()

	if err := a.Run(); err != nil {
		log.Fatalf("app run failed: %v", err)
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
