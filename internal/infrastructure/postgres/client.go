package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/endorsain/nictofilia-go-api/internal/config"
	_ "github.com/lib/pq"
)

type Client struct {
	db *sql.DB
}

func New(ctx context.Context, cfg config.DatabaseConfig, logger *slog.Logger) (*Client, error) {
	logger.Info("connecting to PostgreSQL")

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		logger.Error("failed to connect to PostgreSQL", "error", err.Error())
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		logger.Error("failed to connect to PostgreSQL", "error", err.Error())
		return nil, fmt.Errorf("ping postgres connection: %w", err)
	}

	logger.Info("PostgreSQL connection established")
	return &Client{db: db}, nil
}

func (c *Client) DB() *sql.DB {
	return c.db
}

func (c *Client) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Close()
}
