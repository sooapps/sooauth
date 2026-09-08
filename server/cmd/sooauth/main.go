package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sooapps/sooauth/server/internal/bootstrap"
	"github.com/sooapps/sooauth/server/internal/config"
	"github.com/sooapps/sooauth/server/internal/httpserver"
	"github.com/sooapps/sooauth/server/internal/migrate"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "run migrations and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if *migrateOnly {
		if cfg.DatabaseURL == "" {
			slog.Error("DATABASE_URL required")
			os.Exit(1)
		}
		if err := migrate.Up(cfg.DatabaseURL); err != nil {
			slog.Error("migrate", "err", err)
			os.Exit(1)
		}
		slog.Info("migrations applied")
		return
	}

	ctx := context.Background()
	boot, err := bootstrap.Run(ctx, cfg)
	if err != nil {
		slog.Error("bootstrap", "err", err)
		os.Exit(1)
	}

	srv, err := httpserver.New(cfg, boot.DB, boot.SigningKey)
	if err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("listening", "port", cfg.Port, "env", cfg.Env)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	srv.Close()
}
