package controller

import (
	"log/slog"
	"net/http"

	"github.com/avran02/fileshare/gateway/internal/service"
)

type ShareController interface {
	Share(w http.ResponseWriter, r *http.Request)
	Unshare(w http.ResponseWriter, r *http.Request)
}

type shareController struct {
	service service.ShareService
}

func (c *shareController) Share(w http.ResponseWriter, r *http.Request) {
	slog.Info("Share a file with users")
	c.service.Share()
	_, err := w.Write([]byte("Share a file with users"))
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func (c *shareController) Unshare(w http.ResponseWriter, r *http.Request) {
	slog.Info("Unshare a file with users")
	c.service.Unshare()
	_, err := w.Write([]byte("Unshare a file with users"))
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func NewShareController(service service.ShareService) ShareController {
	return &shareController{
		service: service,
	}
}
