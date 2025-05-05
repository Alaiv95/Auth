package auth

import (
	"Auth/internal/lib/jwt"
	"Auth/internal/lib/logger/sl"
	"Auth/internal/models"
	"Auth/internal/storage"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Storage interface {
	AddUser(ctx context.Context, email string, passwordHash []byte) (userID int64, err error)
	User(ctx context.Context, email string) (user models.User, err error)
}

type AppProvider interface {
	App(ctx context.Context, appId int) (app models.App, err error)
}

type Service struct {
	log         *slog.Logger
	storage     Storage
	appProvider AppProvider
	tokenTTL    time.Duration
}

func New(log *slog.Logger, storage Storage, provider AppProvider, tokenTTL time.Duration) *Service {
	return &Service{
		log:         log,
		storage:     storage,
		appProvider: provider,
		tokenTTL:    tokenTTL,
	}
}

func (s *Service) Register(ctx context.Context, email string, password string) (userID int64, err error) {
	const op = "auth.service.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("Registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.storage.AddUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to add user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Service) Login(ctx context.Context, email string, password string, appId int) (token string, err error) {
	const op = "auth.service.Login"

	log := s.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("Authenticating user")

	user, err := s.storage.User(ctx, email)
	if err != nil {
		log.Warn("failed to find user", sl.Err(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		log.Error("password invalid", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := s.appProvider.App(ctx, appId)
	if err != nil {
		log.Error("failed to find app", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in")

	token, err = jwt.New(user, app, s.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}
