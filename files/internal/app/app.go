package app

import (
	"log/slog"
	"net"
	"os"

	"github.com/avran02/fileshare/files/internal/config"
	"github.com/avran02/fileshare/files/internal/controller"
	"github.com/avran02/fileshare/files/internal/server"
	"github.com/avran02/fileshare/files/internal/service"
	pb "github.com/avran02/fileshare/proto/filespb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var opts []grpc.ServerOption

type App struct {
	Config     *config.Config
	Controller controller.FileServerController
	Server     server.FileServer
	Service    service.FilesService
}

func (app *App) Run() {
	host := ":" + app.Config.Server.Port
	lis, err := net.Listen("tcp", host)
	if err != nil {
		slog.Error("failed to listen:\n" + err.Error())
		os.Exit(1)
	}

	slog.Info("Listening on " + host)

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterFileServiceServer(grpcServer, app.Server)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("fileservice", grpc_health_v1.HealthCheckResponse_SERVING)

	err = grpcServer.Serve(lis)
	if err != nil {
		slog.Error("failed to serve:\n" + err.Error())
		os.Exit(1)
	}
}

func New() *App {
	conf := config.New()
	service := service.New(conf.Minio)
	controller := controller.New(service)
	server := server.New(controller)

	return &App{
		Config:     conf,
		Controller: controller,
		Server:     server,
		Service:    service,
	}
}
