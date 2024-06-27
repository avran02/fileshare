package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/avran02/fileshare/files/internal/config"
	"github.com/avran02/fileshare/files/internal/dto"
	"github.com/avran02/fileshare/files/internal/service"
	"github.com/avran02/fileshare/files/pb"
)

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
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return &pb.ListFilesResponse{Files: files}, nil
}

func (c fileServerController) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	ctx := stream.Context()
	streamErrChan := make(chan error, 1)

	file, err := c.Service.DownloadFile(ctx, req.UserID, req.FilePath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	go c.asyncSendFile(stream, file, streamErrChan)

	if err = <-streamErrChan; err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

func (c fileServerController) UploadFile(stream pb.FileService_UploadFileServer) error {
	slog.Info("Upload file")

	streamErrChan := make(chan error, 1)
	ctx := stream.Context()

	r, err := stream.Recv()
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to receive upload file request: %w", err)
	}

	if len(r.Content) != 0 {
		slog.Warn("Content should be empty")
		return fmt.Errorf("content should be empty")
	}

	requestDTO, err := dto.NewUploadFileStreamRequest(r.UserID, r.FilePath)
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to get upload file request: %w", err)
	}
	defer requestDTO.CloseReader()

	go c.asyncGetFileFromGrpcStream(stream, requestDTO, streamErrChan)

	if err = c.Service.UploadFile(ctx, requestDTO); err != nil {
		err = fmt.Errorf("failed to upload file: %w", err)
		slog.Error(err.Error())
		return err
	}

	if err = <-streamErrChan; err != nil {
		err = fmt.Errorf("failed while uploading file from stream: %w", err)
		slog.Error(err.Error())
		return err
	}

	if err = stream.SendAndClose(&pb.UploadFileResponse{Success: true}); err != nil {
		err = fmt.Errorf("failed to send upload file response: %w", err)
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

func (c fileServerController) asyncSendFile(stream pb.FileService_DownloadFileServer, file io.ReadCloser, streamErrChan chan error) {
	defer close(streamErrChan)
	defer file.Close()
	buf := make([]byte, config.StreamChunkSize)

	for {
		n, err := file.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if n == 0 {
					break
				}
			} else {
				err = fmt.Errorf("failed to read file: %w", err)
				slog.Error(err.Error())
				streamErrChan <- err
			}
		}

		if err = stream.Send(&pb.DownloadFileResponse{
			Content: buf[:n],
		}); err != nil {
			streamErrChan <- fmt.Errorf("failed to send download file response: %w", err)
		}
	}

	if err := stream.Send(&pb.DownloadFileResponse{
		Success: true,
	}); err != nil {
		streamErrChan <- fmt.Errorf("failed to send download file response: %w", err)
	}
}

func (c fileServerController) asyncGetFileFromGrpcStream(stream pb.FileService_UploadFileServer, requestDTO *dto.UploadFileStreamRequest, streamErrChan chan error) {
	defer close(streamErrChan)
	defer requestDTO.CloseWriter()

	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			err = fmt.Errorf("failed to receive upload file request: %w", err)
			slog.Error(err.Error())
			streamErrChan <- err
			return
		}

		_, err = requestDTO.Write(req.Content)
		if err != nil {
			err = fmt.Errorf("failed to write upload file request: %w", err)
			slog.Error(err.Error())
			streamErrChan <- err
			return
		}
	}
}

func New(service service.FilesService) FileServerController {
	return fileServerController{
		Service: service,
	}
}
