package db

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBRepository struct {
	db *sql.DB
}

var table = "shorten_urls"

func CheckConnection(dsn string) error {
	if dsn == "" {
		return fmt.Errorf("data source name is empty")
	}

	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

func New(dsn string) (*DBRepository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DBRepository{db: db}, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (r *DBRepository) Save(originalURL string) (string, error) {
	log.Printf("Saving original URL: %s", originalURL)

	var existingShort string
	checkQuery := fmt.Sprintf(`SELECT short_url FROM %s WHERE original_url = $1`, table)

	err := r.db.QueryRow(checkQuery, originalURL).Scan(&existingShort)
	if err == nil {
		log.Printf("URL already exists with short ID: %s", existingShort)
		return existingShort, nil
	} else if err != sql.ErrNoRows {
		log.Printf("Error checking existing URL: %v", err)
		return "", fmt.Errorf("failed to check existing URL: %w", err)
	}

	shortID, err := randstr.GenerateRandomStringURLSafe(8)
	if err != nil {
		log.Printf("Error generating short ID: %v", err)
		return "", err
	}

	insertQuery := fmt.Sprintf(`INSERT INTO %s (short_url, original_url) VALUES ($1, $2)`, table)

	_, err = r.db.Exec(insertQuery, shortID, originalURL)
	if err != nil {
		log.Printf("Error inserting into database: %v", err)
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	log.Printf("Successfully saved URL with short ID: %s", shortID)
	return shortID, nil
}

func (r *DBRepository) Get(shortURL string) (string, bool) {
	log.Printf("Getting original URL for short URL: %s", shortURL)

	query := fmt.Sprintf(`SELECT original_url FROM %s WHERE short_url = $1`, table)

	var originalURL string
	err := r.db.QueryRow(query, shortURL).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Short URL %s not found", shortURL)
			return "", false
		}
		log.Printf("Error querying database: %v", err)
		return "", false
	}

	log.Printf("Found original URL: %s for short URL: %s", originalURL, shortURL)
	return originalURL, true
}
