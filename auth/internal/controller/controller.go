package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/avran02/fileshare/auth/internal/service"
	pb "github.com/avran02/fileshare/proto/authpb"
	"github.com/google/uuid"
)

type Controller interface {
	Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error)
	Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
	RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error)
	ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error)
	Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error)
}

// implements pb.AuthServiceServer.
type controller struct {
	servcie service.Service
}

func (c *controller) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := c.servcie.Register(uuid.NewString(), req.Username, req.Password); err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	resp := &pb.RegisterResponse{
		Success: true,
	}

	return resp, nil
}

func (c *controller) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	accessToken, refreshToken, err := c.servcie.Login(req.Username, req.Password)
	if err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("failed to login user: %w", err)
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (c *controller) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	token, err := c.servcie.RefreshToken(req.RefreshToken)
	if err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken: token,
	}, nil
}

func (c *controller) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	id, err := c.servcie.ValidateToken(req.AccessToken)
	if err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	return &pb.ValidateTokenResponse{
		Id: id,
	}, nil
}

func (c *controller) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	ok, err := c.servcie.Logout(req.AccessToken)
	if err != nil {
		err = fmt.Errorf("failed to logout: %w", err)
		slog.Info(err.Error())
		return nil, err
	}

	return &pb.LogoutResponse{
		Success: ok,
	}, nil
}

func New(service service.Service) Controller {
	return &controller{
		servcie: service,
	}
}
