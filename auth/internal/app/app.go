package app

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/avran02/fileshare/auth/internal/config"
	"github.com/avran02/fileshare/auth/internal/controller"
	"github.com/avran02/fileshare/auth/internal/pkg/jwt"
	"github.com/avran02/fileshare/auth/internal/repo"
	"github.com/avran02/fileshare/auth/internal/server"
	"github.com/avran02/fileshare/auth/internal/service"
	pb "github.com/avran02/fileshare/proto/authpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type App struct {
	server     *server.Server
	config     *config.Config
	repo       repo.Repo
	jwt        jwt.JwtGenerator
	controller controller.Controller
}

func (app *App) Run() {
	host := ":" + app.config.Server.Port
	lis, err := net.Listen("tcp", host)
	if err != nil {
		slog.Error(fmt.Sprintf("can't listen on %s: \n%s", host, err.Error()))
		os.Exit(1)
	}

	slog.Info("Listening on " + host)
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterAuthServiceServer(grpcServer, app.server)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("authservice", grpc_health_v1.HealthCheckResponse_SERVING)

	err = grpcServer.Serve(lis)
	if err != nil {
		slog.Error(fmt.Sprintf("can't start grpc server: \n%s", err.Error()))
		os.Exit(1)
	}
}

func New() *App {
	config := config.New()
	jwtConf := jwt.New(config.JWT)

	repo := repo.New(&config.DB)
	service := service.New(repo, jwtConf)
	controller := controller.New(service)
	server := server.New(controller)

	return &App{
		config:     config,
		repo:       repo,
		jwt:        jwtConf,
		controller: controller,
		server:     server,
	}
}
