package grpcapp

import (
	"Auth/internal/grpc/auth"
	"context"
	"fmt"
	authv1 "github.com/Alaiv95/Protos/gen/go/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, authService auth.Auth, port int) *App {
	recOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) error {
			log.Error("Recovery from panic", slog.Any("panic", p))

			return status.Error(codes.Internal, "Internal Server Error")
		}),
	}

	logOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadSent, logging.PayloadReceived),
	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recOpts...),
		logging.UnaryServerInterceptor(interceptorLogger(log), logOpts...),
	))

	reflection.Register(server)
	auth.Register(server, authService)

	return &App{
		log:        log,
		gRPCServer: server,
		port:       port,
	}
}

func (a *App) Stop() {
	const op = "grpc.stop"

	a.log.With(slog.String("op", op)).Info(
		"Stopping gRPC server", slog.Int("port", a.port),
	)

	a.gRPCServer.GracefulStop()
}

func (a *App) MustRun() {
	if err := a.run(); err != nil {
		panic(err)
	}
}

func (a *App) run() error {
	const op = "grpc.run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("starting gRPC server on port", slog.String("addr", l.Addr().String()))

	if err = a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		for i, field := range fields {
			switch v := field.(type) {
			case *authv1.LoginRequest:
				fields[i] = struct {
					Email string
					App   int32
				}{
					v.Email,
					v.AppId,
				}
			case *authv1.RegisterRequest:
				fields[i] = v.Email
			}
		}

		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}
