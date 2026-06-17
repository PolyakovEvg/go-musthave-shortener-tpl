package db

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/randstr"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBRepository struct {
	db *sql.DB
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

func (r *DBRepository) Ping() error {
	return r.db.Ping()
}

func (r *DBRepository) Save(userID, originalURL string) (string, error) {
	shortID, err := randstr.GenerateRandomStringURLSafe(8)
	if err != nil {
		return "", err
	}

	var returnedShort string
	query := `
	INSERT INTO shorten_urls (short_url, original_url, user_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (original_url) DO NOTHING
	RETURNING short_url`

	err = r.db.QueryRow(query, shortID, originalURL, userID).Scan(&returnedShort)
	if err != nil {
		if err == sql.ErrNoRows {
			selectQuery := `SELECT short_url FROM shorten_urls WHERE original_url = $1`
			err = r.db.QueryRow(selectQuery, originalURL).Scan(&returnedShort)
			if err != nil {
				return "", fmt.Errorf("failed to fetch existing URL: %w", err)
			}
			return returnedShort, repository.ErrConflict
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	return returnedShort, nil
}

func (r *DBRepository) SaveBatch(userID string, batch []model.BatchRequest) ([]model.BatchResponse, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	responses := make([]model.BatchResponse, 0, len(batch))

	for _, req := range batch {
		var existingShort string
		selectQuery := `SELECT short_url FROM shorten_urls WHERE original_url = $1`
		err := tx.QueryRow(selectQuery, req.OriginalURL).Scan(&existingShort)
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

		insertQuery := `INSERT INTO shorten_urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`
		_, err = tx.Exec(insertQuery, shortID, req.OriginalURL, userID)
		if err != nil {
			return nil, fmt.Errorf("error inserting into database: %w", err)
		}

		responses = append(responses, model.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortID,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return responses, nil
}

func (r *DBRepository) Get(shortURL string) (*model.URL, bool) {
	var rec model.URL
	query := `SELECT short_url, original_url, user_id, is_deleted FROM shorten_urls WHERE short_url=$1`
	err := r.db.QueryRow(query, shortURL).Scan(&rec.ShortURL, &rec.OriginalURL, &rec.UserID, &rec.IsDeleted)
	if err != nil {
		return nil, false
	}

	return &rec, true
}

func (r *DBRepository) GetByUser(userID string) ([]model.URL, error) {
	query := `SELECT short_url, original_url FROM shorten_urls WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.URL
	for rows.Next() {
		var u model.URL
		u.UserID = userID
		if err := rows.Scan(&u.ShortURL, &u.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}

func (r *DBRepository) MarkDeleted(userID string, shorts []string) error {
	if len(shorts) == 0 {
		return nil
	}

	query := `UPDATE shorten_urls SET is_deleted = TRUE WHERE user_id = $1 AND short_url = ANY($2)`
	_, err := r.db.Exec(query, userID, shorts)
	return err
}

// Close закрывает соединение с базой данных.
func (r *DBRepository) Close() error {
	return r.db.Close()
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
