package app

import (
	"context"
	"log"
	"log/slog"

	"github.com/avran02/fileshare/files/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func connectToFilesServer(endpoint string) (pb.FileServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to connect to gRPC server: ", err)
	}
	status := chechFilesServerHealth(conn)
	slog.Info("Connected to gRPC server at http://" + endpoint + "\ngRPC health status: " + status)
	return pb.NewFileServiceClient(conn), conn
}

func chechFilesServerHealth(conn *grpc.ClientConn) string {
	resp, err := grpc_health_v1.NewHealthClient(conn).Check(
		context.Background(),
		&grpc_health_v1.HealthCheckRequest{
			Service: "fileservice",
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
