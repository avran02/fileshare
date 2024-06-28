package service

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/avran02/fileshare/proto/authpb"
)

type UserService interface {
	RegisterUser(ctx context.Context, username, password string) (bool, error)
	LoginUser(ctx context.Context, username, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, accessToken string) (bool, error)

	GetGrpcClient() pb.AuthServiceClient
}

type userService struct {
	authServiceClient pb.AuthServiceClient
}

func (s *userService) RegisterUser(ctx context.Context, username, password string) (bool, error) {
	resp, err := s.authServiceClient.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return false, err
	}

	return resp.Success, nil
}

func (s *userService) LoginUser(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	resp, err := s.authServiceClient.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		slog.Error(err.Error())
		return "", "", fmt.Errorf("failed to login: %w", err)
	}

	return resp.AccessToken, resp.RefreshToken, nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	resp, err := s.authServiceClient.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		slog.Error(err.Error())
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	return resp.AccessToken, nil
}

func (s *userService) Logout(ctx context.Context, accessToken string) (bool, error) {
	resp, err := s.authServiceClient.Logout(ctx, &pb.LogoutRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		slog.Error(err.Error())
		return false, fmt.Errorf("failed to logout: %w", err)
	}

	return resp.Success, nil
}

func (s *userService) GetGrpcClient() pb.AuthServiceClient {
	return s.authServiceClient
}

func NewUserService(client pb.AuthServiceClient) UserService {
	return &userService{
		authServiceClient: client,
	}
}
