package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/avran02/fileshare/auth/internal/pkg/jwt"
	"github.com/avran02/fileshare/auth/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("user already exists")
	ErrTokenDoesntExist = errors.New("token doesn't exist")
)

type Service interface {
	Register(id, username, password string) error
	Login(username, password string) (accessToken, refreshToken string, err error)
	RefreshToken(token string) (string, error)
	ValidateToken(token string) (string, error)
	Logout(token string) (bool, error)
}

type service struct {
	repo repo.Repo
	jwt  jwt.JwtGenerator
}

func (s *service) Register(id, username, password string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		err = fmt.Errorf("failed to hash password: %w", err)
		slog.Error(err.Error())
		return err
	}

	if err = s.repo.CreateUser(username, string(hashedPass)); err != nil {
		err = fmt.Errorf("failed to create user: %w", err)
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (s *service) Login(username, password string) (accessToken, refreshToken string, err error) {
	user, err := s.repo.FindUserByUsername(username)
	if err != nil {
		err = fmt.Errorf("failed to find user: %w", err)
		slog.Error(err.Error())
		return "", "", err
	}

	slog.Info("Found user: " + user.Username)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		err = fmt.Errorf("failed to compare password: %w", err)
		slog.Error(err.Error())
		return "", "", err
	}

	accessTokenModel, err := s.jwt.Generate(user.ID, true)
	if err != nil {
		err = fmt.Errorf("failed to generate token: %w", err)
		slog.Error(err.Error())
		return "", "", err
	}

	refreshTokenModel, err := s.jwt.Generate(user.ID, false)
	if err != nil {
		err = fmt.Errorf("failed to generate token: %w", err)
		slog.Error(err.Error())
		return "", "", err
	}

	if err = s.repo.DeleteUserTokensAndWriteNew(user.ID, accessTokenModel, refreshTokenModel); err != nil {
		err = fmt.Errorf("failed to delete tokens: %w", err)
		slog.Error(err.Error())
		return "", "", err
	}

	return accessTokenModel.Token, refreshTokenModel.Token, nil
}

func (s *service) RefreshToken(token string) (string, error) {
	userId, isAccessToken, err := s.jwt.Validate(token) //nolint
	if err != nil {
		err = fmt.Errorf("failed to validate token: %w", err)
		slog.Error(err.Error())
		return "", err
	}

	if isAccessToken {
		slog.Error("expected refresh token, got access token: " + jwt.ErrWrongTokenType.Error())
		return "", fmt.Errorf("expected refresh token, got access token: %w", jwt.ErrWrongTokenType)
	}

	tokenExists, err := s.repo.CheckTokenExists(token)
	if err != nil {
		err = fmt.Errorf("failed to check token exists: %w", err)
		slog.Error(err.Error())
		return "", err
	}

	if !tokenExists {
		slog.Error("token " + token + " does not exist")
		return "", fmt.Errorf("token %s does not exist: %w", token, ErrTokenDoesntExist)
	}

	newAccessToken, err := s.jwt.Generate(userId, true)
	if err != nil {
		err = fmt.Errorf("failed to generate token: %w", err)
		slog.Error(err.Error())
		return "", err
	}

	if err = s.repo.ReplaceUserAccessToken(newAccessToken); err != nil {
		err = fmt.Errorf("failed to write token: %w", err)
		slog.Error(err.Error())
		return "", err
	}

	return newAccessToken.Token, nil
}

func (s *service) ValidateToken(token string) (string, error) {
	userID, isAccessToken, err := s.jwt.Validate(token)
	if err != nil {
		err = fmt.Errorf("failed to validate token: %w", err)
		slog.Error(err.Error())
		return "", err
	}

	if !isAccessToken {
		slog.Error("expected access token, got refresh token: " + jwt.ErrWrongTokenType.Error())
		return "", fmt.Errorf("expected access token, got refresh token: %w", jwt.ErrWrongTokenType)
	}

	exists, err := s.repo.CheckTokenExists(token)
	if err != nil {
		err = fmt.Errorf("failed to check token exists: %w", err)
		slog.Error(err.Error())
		return "", err
	}

	if !exists {
		slog.Error("token " + token + " does not exist")
		return "", fmt.Errorf("token %s does not exist: %w", token, ErrTokenDoesntExist)
	}

	return userID, nil
}

func (s *service) Logout(token string) (bool, error) {
	userID, isAccessToken, err := s.jwt.Validate(token)
	if err != nil {
		err = fmt.Errorf("failed to validate token: %w", err)
		slog.Error(err.Error())
		return false, err
	}

	if !isAccessToken {
		slog.Error("expected access token, got refresh token: " + jwt.ErrWrongTokenType.Error())
		return false, fmt.Errorf("expected access token, got refresh token: %w", jwt.ErrWrongTokenType)
	}

	exists, err := s.repo.CheckTokenExists(token)
	if err != nil {
		err = fmt.Errorf("failed to check token exists: %w", err)
		slog.Error(err.Error())
		return false, err
	}

	if !exists {
		slog.Error("token " + token + " does not exist")
		return false, fmt.Errorf("token %s does not exist: %w", token, ErrTokenDoesntExist)
	}

	err = s.repo.DeleteAllUserTokens(userID)
	if err != nil {
		err = fmt.Errorf("failed to remove token: %w", err)
		slog.Error(err.Error())
		return false, err
	}

	return true, nil
}

func New(repo repo.Repo, jwt jwt.JwtGenerator) Service {
	return &service{
		repo: repo,
		jwt:  jwt,
	}
}
