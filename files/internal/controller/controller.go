package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/avran02/fileshare/files/internal/config"
	"github.com/avran02/fileshare/files/internal/service"
	"github.com/avran02/fileshare/files/pb"
)

// var timeOut = time.Duration(999999999999999999) //nolint

type FileServerController interface {
	ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error)
	DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error
	UploadFile(stream pb.FileService_UploadFileServer) error
	RemoveFile(ctx context.Context, req *pb.RemoveFileRequest) (*pb.RemoveFileResponse, error)
	RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error)
}

type fileServerController struct {
	Service service.FilesService
}

func (c fileServerController) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	files, err := c.Service.ListFiles(ctx, req.UserID, req.FilePath)
	if err != nil {
		return nil, err
	}
	slog.Info("List files: " + fmt.Sprint(files))
	return &pb.ListFilesResponse{Files: files}, nil
}

func (c fileServerController) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	ctx := context.Background()

	object, err := c.Service.DownloadFile(ctx, req.UserID, req.FilePath)
	if err != nil {
		return err
	}

	defer object.Close()

	buf := make([]byte, config.StreamChunkSize)

	for {
		n, err := object.Read(buf)
		if err != nil && err.Error() != "EOF" {
			slog.Error(err.Error())
			return err
		}

		err = stream.Send(&pb.DownloadFileResponse{
			Content: buf[:n],
		})
		if err != nil {
			return err
		}
	}
}

func (c fileServerController) UploadFile(stream pb.FileService_UploadFileServer) error {
	ctx := context.Background()
	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			slog.Error(err.Error())
			return err
		}

		// TODO: Оптимизировать это так как сейчас он отправляется хуй пойми как
		err = c.Service.UploadFile(ctx, req.Content, req.UserID, req.FilePath)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
	}

	err := stream.SendAndClose(&pb.UploadFileResponse{
		Success: true,
	})
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (c fileServerController) RemoveFile(ctx context.Context, req *pb.RemoveFileRequest) (*pb.RemoveFileResponse, error) {
	err := c.Service.RemoveFile(ctx, req.UserID, req.FilePath)
	if err != nil {
		return &pb.RemoveFileResponse{
			Success: false,
		}, err
	}

	return &pb.RemoveFileResponse{
		Success: true,
	}, nil
}

func (c fileServerController) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	err := c.Service.RegisterUser(ctx, req.UserID)
	if err != nil {
		return &pb.RegisterUserResponse{
			Success: false,
		}, err
	}

	return &pb.RegisterUserResponse{
		Success: true,
	}, nil
}

func New(service service.FilesService) FileServerController {
	return fileServerController{
		Service: service,
	}
}
