package db

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
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

const tableName = "shorten_urls"

const (
	insertQuery = `INSERT INTO shorten_urls (short_url, original_url) 
                   VALUES ($1, $2) 
                   ON CONFLICT (original_url) DO NOTHING
                   RETURNING short_url`

	selectByOriginalQuery = `SELECT short_url FROM shorten_urls WHERE original_url = $1`

	selectByShortQuery = `SELECT original_url FROM shorten_urls WHERE short_url = $1`

	insertBatchQuery = `INSERT INTO shorten_urls (short_url, original_url) VALUES ($1, $2)`
)

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

func (r *DBRepository) Ping() error {
	return r.db.Ping()
}

func (r *DBRepository) Save(originalURL string) (string, error) {
	log.Printf("Saving original URL: %s", originalURL)

	shortID, err := randstr.GenerateRandomStringURLSafe(8)
	if err != nil {
		return "", err
	}

	var returnedShort string
	err = r.db.QueryRow(insertQuery, shortID, originalURL).Scan(&returnedShort)

	if err != nil {
		if err == sql.ErrNoRows {
			err = r.db.QueryRow(selectByOriginalQuery, originalURL).Scan(&returnedShort)
			if err != nil {
				return "", fmt.Errorf("failed to fetch existing URL: %w", err)
			}
			return returnedShort, repository.ErrConflict
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	return returnedShort, nil
}

func (r *DBRepository) SaveBatch(batch []model.BatchRequest) ([]model.BatchResponse, error) {
	log.Printf("Saving batch of %d URLs", len(batch))

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	responses := make([]model.BatchResponse, 0, len(batch))

	for _, req := range batch {
		var existingShort string
		err := tx.QueryRow(selectByOriginalQuery, req.OriginalURL).Scan(&existingShort)
		if err == nil {
			responses = append(responses, model.BatchResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      existingShort,
			})
			continue
		} else if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check existing URL: %w", err)
		}

		shortID, err := randstr.GenerateRandomStringURLSafe(8)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(insertBatchQuery, shortID, req.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("Error inserting into database: %w", err)
		}

		responses = append(responses, model.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortID,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("Error committing transaction: %w", err)
	}

	log.Printf("Successfully saved batch of %d URLs", len(batch))
	return responses, nil
}

func (r *DBRepository) Get(shortURL string) (string, bool) {
	log.Printf("Getting original URL for short URL: %s", shortURL)

	var originalURL string
	err := r.db.QueryRow(selectByShortQuery, shortURL).Scan(&originalURL)
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
