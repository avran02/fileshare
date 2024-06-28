package controller

import (
	"log/slog"
	"net/http"

	"github.com/avran02/fileshare/gateway/internal/dto"
	"github.com/avran02/fileshare/gateway/internal/service"
	pb "github.com/avran02/fileshare/proto/authpb"
)

type UsersController interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)

	GetGrpcClient() pb.AuthServiceClient
}

type userController struct {
	service service.UserService
}

func (c *userController) Login(w http.ResponseWriter, r *http.Request) {
	slog.Info("Login a user")
	ctx := r.Context()

	var req dto.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := c.service.LoginUser(ctx, req.Username, req.Password)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *userController) Register(w http.ResponseWriter, r *http.Request) {
	slog.Info("Register a new user")
	ctx := r.Context()

	var req dto.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok, err := c.service.RegisterUser(ctx, req.Username, req.Password)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.RegisterUserResponse{
		Success: ok,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *userController) RefreshToken(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	slog.Info("Update user token")
	ctx := r.Context()

	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken, err := c.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := dto.RefreshTokenResponse{
		AccessToken: accessToken,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *userController) Logout(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	slog.Info("Logout a user")
	ctx := r.Context()

	var req dto.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok, err := c.service.Logout(ctx, req.AccessToken)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.LogoutResponse{
		Success: ok,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *userController) GetGrpcClient() pb.AuthServiceClient {
	return c.service.GetGrpcClient()
}

func NewUsersController(service service.UserService) UsersController {
	return &userController{
		service: service,
	}
}
