package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/config"
	"github.com/sooapps/sooauth/server/internal/crypto/signing"
	"github.com/sooapps/sooauth/server/internal/migrate"
)

type Result struct {
	DB         *pgxpool.Pool
	SigningKey *signing.Key
}

func Run(ctx context.Context, cfg config.Config) (*Result, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required for bootstrap")
	}

	if cfg.AutoMigrate {
		slog.Info("running migrations")
		if err := migrate.Up(cfg.DatabaseURL); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}

	key, err := signing.EnsureActiveKey(ctx, db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("signing key: %w", err)
	}
	slog.Info("signing key ready", "kid", key.KID)

	return &Result{DB: db, SigningKey: key}, nil
}
