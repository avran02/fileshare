package app

import (
	"context"
	"log"
	"log/slog"

	authpb "github.com/avran02/fileshare/proto/authpb"
	filespb "github.com/avran02/fileshare/proto/filespb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func connectToFilesServer(endpoint string) (filespb.FileServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to connect to gRPC server: ", err)
	}
	status := checkServerHealth(conn, "fileservice")
	slog.Info("Connected to gRPC server at http://" + endpoint + "\ngRPC health status: " + status)
	return filespb.NewFileServiceClient(conn), conn
}

func connectToAuthService(endpoint string) (authpb.AuthServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to connect to gRPC server: ", err)
	}
	status := checkServerHealth(conn, "authservice")
	slog.Info("Connected to gRPC server at http://" + endpoint + "\ngRPC health status: " + status)
	return authpb.NewAuthServiceClient(conn), conn
}

func checkServerHealth(conn *grpc.ClientConn, serviceName string) string {
	resp, err := grpc_health_v1.NewHealthClient(conn).Check(
		context.Background(),
		&grpc_health_v1.HealthCheckRequest{
			Service: serviceName,
		},
	)
	if err != nil {
		log.Fatal("Failed to connect to gRPC server: ", err)
	}
	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		log.Fatal("Failed to connect to gRPC server: ", resp.GetStatus())
	}

	return resp.Status.String()
}
