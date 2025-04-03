package service

import (
	"context"
	"github.com/go-playground/validator"
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
	logger    *zap.SugaredLogger
	cfg       config.AuthConfig
	repo      repo.Repository
	validator *validator.Validate
}

func NewServerAuth(cfg config.AuthConfig, repo repo.Repository, logger *zap.SugaredLogger) *ServerAuth {
	return &ServerAuth{
		logger:    logger,
		cfg:       cfg,
		repo:      repo,
		validator: validator.New(),
	}
}

type RegisterRequestValidation struct {
	Email     string `validate:"required,email"`
	Username  string `validate:"required,min=3,max=30"`
	Password  string `validate:"required,min=6,max=50"`
	FirstName string `validate:"required,alpha,min=2,max=30"`
	LastName  string `validate:"required,alpha,min=2,max=30"`
}

type LoginRequestValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6,max=50"`
}

func (s *ServerAuth) Register(ctx context.Context, req *gen.RegisterRequest) (*gen.RegisterResponse, error) {

	input := RegisterRequestValidation{
		Email:     req.GetEmail(),
		Username:  req.GetUsername(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	}

	if err := s.validator.Struct(input); err != nil {
		s.logger.Errorf("Invalid input: %v", err)
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}

	exists, err := s.repo.CheckUserExists(ctx, input.Username, input.Email)
	if err != nil {
		s.logger.Errorf("Error checking user existence: %v", err)
		return nil, status.Error(codes.Internal, "error checking user")
	}
	if exists {
		s.logger.Errorf("User %s already exists", input.Username)
		return nil, status.Error(codes.AlreadyExists, "User already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "error hashing password")
	}

	resp, err := s.repo.RegisterUser(ctx, &repo.User{
		Email:     input.Email,
		Username:  input.Username,
		Password:  string(passwordHash),
		FirstName: input.FirstName,
		LastName:  input.LastName,
	})
	if err != nil {
		s.logger.Errorf("Failed to register user: %v", err)
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &gen.RegisterResponse{Message: resp}, nil
}

func (s *ServerAuth) Login(ctx context.Context, req *gen.LoginRequest) (*gen.LoginResponse, error) {

	input := LoginRequestValidation{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	if err := s.validator.Struct(input); err != nil {
		s.logger.Errorf("Invalid input: %v", err)
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}

	s.logger.Infof("Login attempt for user: %s", input.Email)

	user, err := s.repo.GetUserByUsername(ctx, input.Email)
	if err != nil {
		s.logger.Errorf("User not found: %v", err)
		return nil, status.Error(codes.NotFound, "User not found")
	}

	s.logger.Infof("User found: %s", input.Email)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		s.logger.Errorf("Invalid password")
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	s.logger.Infof("User successfully logged in: %s", input.Email)

	return &gen.LoginResponse{Token: "JWT_TOKEN"}, nil
}
