package controller

import (
	"log/slog"
	"net/http"

	"github.com/avran02/fileshare/gateway/internal/service"
)

type UsersController interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type userController struct {
	service service.UserService
}

func (c *userController) Login(w http.ResponseWriter, r *http.Request) {
	slog.Info("Login a user")
	c.service.LoginUser()
	_, err := w.Write([]byte("Login a user"))
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func (c *userController) Register(w http.ResponseWriter, r *http.Request) {
	slog.Info("Register a new user")
	c.service.RegisterUser()
	_, err := w.Write([]byte("Register a new user"))
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func (c *userController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	slog.Info("Update user token")
	c.service.RefreshToken()
	_, err := w.Write([]byte("Update user token"))
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func (c *userController) Logout(w http.ResponseWriter, r *http.Request) {
	slog.Info("Logout a user")
	c.service.Logout()
	_, err := w.Write([]byte("Logout a user"))
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func NewUsersController(service service.UserService) UsersController {
	return &userController{
		service: service,
	}
}
