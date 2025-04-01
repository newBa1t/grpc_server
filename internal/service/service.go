package service

import (
	"context"
	"github.com/newBa1t/grpc_server.git/internal/config"
	"github.com/newBa1t/grpc_server.git/internal/repo"
	"github.com/newBa1t/grpc_server.git/protos/gen"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerAuth struct {
	gen.UnimplementedAuthServiceServer
	logger *zap.SugaredLogger
	cfg    config.AuthConfig
	repo   repo.Repository
}

func NewServerAuth(cfg config.AuthConfig, repo repo.Repository, logger *zap.SugaredLogger) *ServerAuth {
	return &ServerAuth{
		logger: logger,
		cfg:    cfg,
		repo:   repo,
	}
}

type User struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
}

func (s *ServerAuth) Register(ctx context.Context, req *gen.RegisterRequest) (*gen.RegisterResponse, error) {

	exists, err := s.repo.CheckUserExists(ctx, req.GetUsername(), req.GetEmail())
	if err != nil {
		s.logger.Errorf("Username or Email is invalid")
		return nil, status.Error(codes.Internal, "error checking user")
	}
	if exists {
		s.logger.Errorf("User: %s already exists", req.GetUsername())
		return nil, status.Error(codes.AlreadyExists, "User already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), 10)
	if err != nil {
		return nil, status.Error(codes.Internal, "error hashing password")
	}
	req.Password = string(passwordHash)

	resp, err := s.repo.RegisterUser(ctx, &repo.User{
		Email:     req.GetEmail(),
		Username:  req.GetUsername(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	})
	if err != nil {
		s.logger.Errorf("failed to register user: %v", err)
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &gen.RegisterResponse{Message: resp}, nil
}

func (s *ServerAuth) Login(ctx context.Context, req *gen.LoginRequest) (*gen.LoginResponse, error) {

	user, err := s.repo.GetUserByUsername(ctx, req.GetEmail())
	if err != nil {
		s.logger.Errorf("User not found: %s", err)
		return nil, status.Error(codes.NotFound, "User not found")
	}

	checkPasswordHash := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetPassword()))
	if checkPasswordHash != nil {
		s.logger.Errorf("Password is invalid")
		return nil, status.Error(codes.Internal, "password is invalid")
	}

	return &gen.LoginResponse{Token: "JWT_TOKEN"}, nil
}
