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
	slog.Info("List files: " + fmt.Sprint(files))
	return &pb.ListFilesResponse{Files: files}, nil
}

func (c fileServerController) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	ctx := stream.Context()
	streamErrChan := make(chan error, 1)

	file, err := c.Service.DownloadFile(ctx, req.UserID, req.FilePath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer file.Close()

	slog.Info("got file from minio")

	buf := make([]byte, config.StreamChunkSize)

	go func() {
		slog.Info("Start stream")
		defer close(streamErrChan)

		for {
			n, err := file.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if n == 0 {
						slog.Info("End of file")
						break
					}
				} else {
					slog.Error(err.Error())
					streamErrChan <- fmt.Errorf("failed to read file: %w", err)
				}
			}

			slog.Info("read " + fmt.Sprint(n) + " bytes")

			err = stream.Send(&pb.DownloadFileResponse{
				Content: buf[:n],
			})
			if err != nil {
				streamErrChan <- fmt.Errorf("failed to send download file response: %w", err)
			}
			slog.Info("bytes sent")
		}

		slog.Info("End loop")
		err = stream.Send(&pb.DownloadFileResponse{
			Success: true,
		})
		if err != nil {
			streamErrChan <- fmt.Errorf("failed to send download file response: %w", err)
		}

		slog.Info("stream ended")
	}()
	err = <-streamErrChan
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to download file: %w", err)
	}

	slog.Info("Downloaded file")

	return nil
}

func (c fileServerController) UploadFile(stream pb.FileService_UploadFileServer) error {
	slog.Info("Upload file")
	requestDTO := dto.UploadFileStreamRequest{}
	streamErrChan := make(chan error, 1)
	ctx := stream.Context()

	r, err := stream.Recv()
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to receive upload file request: %w", err)
	}

	err = requestDTO.GetData(r.UserID, r.FilePath)
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to get upload file request: %w", err)
	}
	defer requestDTO.CloseReader()

	go func() {
		defer close(streamErrChan)
		defer requestDTO.CloseWriter()

		for {
			slog.Info("Waiting for upload file request")
			req, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					slog.Info("EOF")
					break
				}

				slog.Error(err.Error())
				streamErrChan <- fmt.Errorf("failed to receive upload file request: %w", err)
				return
			}

			_, err = requestDTO.Write(req.Content)
			if err != nil {
				slog.Error(err.Error())
				streamErrChan <- fmt.Errorf("failed to write upload file request: %w", err)
				return
			}
		}
	}()

	slog.Info("Starting upload file")
	_, err = c.Service.UploadFile(ctx, &requestDTO)
	if err != nil {
		slog.Error(fmt.Errorf("failed to upload file: %w", err).Error())
		return fmt.Errorf("failed to upload file: %w", err)
	}

	err = <-streamErrChan
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	err = stream.SendAndClose(&pb.UploadFileResponse{Success: true})
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to send upload file response: %w", err)
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
