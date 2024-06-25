package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/avran02/fileshare/files/pb"
)

var chankSize = 1024 * 1024

type FilesService interface {
	ListFiles(ctx context.Context, userID, filePath string) ([]*pb.FileInfo, error)
	UploadFile(ctx context.Context, reader io.Reader, userID, filePath string) (bool, error)
	DownloadFile()
	RemoveFile(userID, filePath string) (bool, error)
}

type filesService struct {
	filesServerClient pb.FileServiceClient
}

func (s *filesService) ListFiles(ctx context.Context, userID, filePath string) ([]*pb.FileInfo, error) {
	resp, err := s.filesServerClient.ListFiles(ctx, &pb.ListFilesRequest{
		UserID:   userID,
		FilePath: filePath,
	})
	if err != nil {
		slog.Error(err.Error())
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	slog.Info("resp in service.ListFiles: " + fmt.Sprint(resp.Files))
	return resp.Files, nil
}

func (s *filesService) UploadFile(ctx context.Context, reader io.Reader, userID, filePath string) (bool, error) {
	stream, err := s.filesServerClient.UploadFile(ctx)
	if err != nil {
		slog.Error(err.Error())
		return false, fmt.Errorf("failed to create upload stream: %w", err)
	}

	err = stream.Send(&pb.UploadFileRequest{
		UserID:   userID,
		FilePath: filePath,
	})
	if err != nil {
		slog.Error(err.Error())
		return false, fmt.Errorf("failed to send initial request: %w", err)
	}

	slog.Info("Stream started\nSent file path and user id: " + filePath + " " + userID)
	buf := make([]byte, chankSize)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			slog.Error(err.Error())
			return false, fmt.Errorf("failed to read file: %w", err)
		}

		err = stream.Send(&pb.UploadFileRequest{
			Content: buf[:n],
		})
		if err != nil {
			slog.Error(err.Error())
			return false, fmt.Errorf("failed to send file chunk: %w", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		slog.Error(err.Error())
		return false, fmt.Errorf("failed to close and receive response: %w", err)
	}

	slog.Info("Stream closed and response received: " + fmt.Sprint(resp.Success))
	return resp.Success, nil
}

func (s *filesService) DownloadFile() {
	// TODO: add gRPC call
}

func (s *filesService) RemoveFile(userID, filePath string) (bool, error) {
	resp, err := s.filesServerClient.RemoveFile(context.TODO(), &pb.RemoveFileRequest{
		UserID:   userID,
		FilePath: filePath,
	})
	if err != nil {
		slog.Error(err.Error())
		return false, err
	}
	return resp.Success, nil
}

func NewFilesService(pbClient pb.FileServiceClient) FilesService {
	return &filesService{
		filesServerClient: pbClient,
	}
}
