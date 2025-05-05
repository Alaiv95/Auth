package main

import (
	"Auth/internal/config"
	"Auth/internal/lib/logger/sl"
	"Auth/internal/services/auth"
	"Auth/internal/storage/pg"
	"context"
	"google.golang.org/grpc"
	"log/slog"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// конфиг
	cfg := config.MustLoad()
	// логгер
	log := setupLogger(cfg.Env)
	// бд
	storage, err := pg.New(ctx, log, cfg.Dsn)
	if err != nil {
		log.Error("Unable to connect to database", sl.Err(err))
		os.Exit(1)
	}
	// app
	app := auth.New(log, storage, storage, cfg.TokenTtl)
	// запуск grpc
	grpc.NewServer()
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
