package app

import (
	"log/slog"
	"net/http"

	"github.com/avran02/fileshare/gateway/internal/config"
	"github.com/avran02/fileshare/gateway/internal/controller"
	"github.com/avran02/fileshare/gateway/internal/router"
	"github.com/avran02/fileshare/gateway/internal/service"
	"google.golang.org/grpc"
)

type App struct {
	*router.Router
	*config.Config
	*grpc.ClientConn
}

func New() *App {
	conf := config.New()
	client, conn := connectToFilesServer(conf.FileService.Endpoint)

	services := service.Services{
		UserService:  service.NewUserService(),
		FilesService: service.NewFilesService(client),
		ShareService: service.NewShareService(),
	}

	controllers := controller.Controllers{
		UsersController: controller.NewUsersController(services.UserService),
		FilesController: controller.NewFilesController(services.FilesService),
		ShareController: controller.NewShareController(services.ShareService),
	}
	return &App{
		Config:     conf,
		Router:     router.New(controllers),
		ClientConn: conn,
	}
}

func (a *App) RunServer() error {
	endpoint := a.getServerEndpoint()
	slog.Info("Server running on http://" + endpoint)
	s := http.Server{ //nolint
		Addr:    endpoint,
		Handler: a.Router,
	}

	s.RegisterOnShutdown(func() {
		a.ClientConn.Close()
	})

	return s.ListenAndServe()
}

func (a *App) getServerEndpoint() string {
	return a.Server.Host + ":" + a.Server.Port
}
