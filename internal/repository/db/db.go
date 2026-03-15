package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewConnection(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("data source name is empty")
	}

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

	return db, nil
}

func CheckConnection(dsn string) error {
	db, err := NewConnection(dsn)
	if err != nil {
		return err
	}
	db.Close()
	return nil
}
