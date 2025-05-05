package pg

import (
	"Auth/internal/lib/logger/sl"
	"Auth/internal/models"
	"Auth/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type Storage struct {
	log *slog.Logger
	db  *pgxpool.Pool
}

func New(ctx context.Context, log *slog.Logger, dsn string) (*Storage, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Error("Unable to connect to database", sl.Err(err))
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		log.Error("Unable to ping database", sl.Err(err))
		return nil, err
	}

	return &Storage{
		log: log,
		db:  conn,
	}, nil
}

func (s *Storage) AddUser(ctx context.Context, email string, passwordHash []byte) (userID int64, err error) {
	const op = "pg.AddUser"

	c, err := s.db.Exec(ctx, "INSERT INTO users(email, passwordHash) VALUES ($1, $2)", email, string(passwordHash))
	if err != nil {
		s.log.Error("Unable to add user", sl.Err(err))

		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23505" {
			return 0, storage.ErrUserExists
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info(fmt.Sprintf("Addes user affected rows: %d", c.RowsAffected()))

	return userID, nil
}

func (s *Storage) User(ctx context.Context, email string) (user models.User, err error) {
	const op = "pg.User"

	err = s.db.QueryRow(ctx, "SELECT id, email, passwordHash FROM users WHERE email = $1", email).Scan(&user.Id, &user.Email, &user.PasswordHash)
	if err != nil {
		s.log.Error("Unable to get user", sl.Err(err))

		if errors.Is(err, pgx.ErrNoRows) {
			return user, storage.ErrUserNotFound
		}

		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, appId int) (app models.App, err error) {
	const op = "pg.App"

	err = s.db.QueryRow(ctx, "SELECT id, name, secret FROM apps WHERE id = $1", appId).Scan(&app.Id, &app.Name, &app.Secret)
	if err != nil {
		s.log.Error("Unable to get app", sl.Err(err))

		if errors.Is(err, pgx.ErrNoRows) {
			return app, storage.ErrAppNotFound
		}

		return app, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
