package auth

import (
	"Auth/internal/services/auth"
	"context"
	"errors"
	authv1 "github.com/Alaiv95/Protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	authv1.UnimplementedAuthServer
	auth Auth
}

type Auth interface {
	Register(ctx context.Context, email string, password string) (userID int64, err error)
	Login(ctx context.Context, email string, password string, appId int) (token string, err error)
}

func Register(server *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(server, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	in *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	if in.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}

	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}

	if in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "AppId is required")
	}

	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authv1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	in *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	if in.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}

	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}

	id, err := s.auth.Register(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "User already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authv1.RegisterResponse{UserId: id}, nil
}
