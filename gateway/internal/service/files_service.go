package service

import (
	"context"
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
		return false, err
	}
	buf := make([]byte, chankSize)
	for {
		_, err := reader.Read(buf)
		if err != nil && err.Error() != "EOF" {
			slog.Error(err.Error())
			return false, err
		}
		err = stream.Send(&pb.UploadFileRequest{
			Content: buf,
		})
		if err != nil {
			slog.Error(err.Error())
			return false, err
		}
		if err == io.EOF {
			break
		}
	}
	err = stream.Send(&pb.UploadFileRequest{
		UserID:   userID,
		FilePath: filePath,
	})
	if err != nil {
		slog.Error(err.Error())
		return false, err
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		slog.Error(err.Error())
		return false, err
	}
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
