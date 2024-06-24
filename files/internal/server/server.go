package server

import (
	"context"

	"github.com/avran02/fileshare/files/internal/controller"
	"github.com/avran02/fileshare/files/pb"
)

type FileServer struct {
	pb.UnimplementedFileServiceServer
	controller.FileServerController
}

func (s FileServer) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	return s.FileServerController.ListFiles(ctx, req)
}

func (s FileServer) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	return s.FileServerController.DownloadFile(req, stream)
}

func (s FileServer) UploadFile(stream pb.FileService_UploadFileServer) error {
	return s.FileServerController.UploadFile(stream)
}

func (s FileServer) RemoveFile(ctx context.Context, req *pb.RemoveFileRequest) (*pb.RemoveFileResponse, error) {
	return s.FileServerController.RemoveFile(ctx, req)
}

func (s FileServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	return s.FileServerController.RegisterUser(ctx, req)
}

func New(controller controller.FileServerController) FileServer {
	return FileServer{
		UnimplementedFileServiceServer: pb.UnimplementedFileServiceServer{},
		FileServerController:           controller,
	}
}
