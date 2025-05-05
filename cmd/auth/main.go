package main

import (
	grpcapp "Auth/internal/app/grpc"
	"Auth/internal/config"
	"Auth/internal/lib/logger/sl"
	"Auth/internal/services/auth"
	"Auth/internal/storage/pg"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	storage, err := pg.New(ctx, log, cfg.Dsn)
	if err != nil {
		log.Error("Unable to connect to database", sl.Err(err))
		os.Exit(1)
	}

	authService := auth.New(log, storage, storage, cfg.TokenTtl)
	app := grpcapp.New(log, authService, cfg.GRPC.Port)

	go func() {
		app.MustRun()
	}()

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	app.Stop()
	storage.Close()
	log.Info("Application gracefully shutting down...")
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case "local":
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return logger
}
