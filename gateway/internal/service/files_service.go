package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	pb "github.com/avran02/fileshare/proto/filespb"
)

var chankSize = 1024 * 1024

type FilesService interface {
	ListFiles(ctx context.Context, userID, filePath string) ([]*pb.FileInfo, error)
	UploadFile(ctx context.Context, reader io.Reader, userID, filePath string) (bool, error)
	DownloadFile(ctx context.Context, userID, filePath string, w *io.PipeWriter) error
	RemoveFile(ctx context.Context, userID, filePath string) (bool, error)
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

	if err = stream.Send(&pb.UploadFileRequest{
		UserID:   userID,
		FilePath: filePath,
	}); err != nil {
		slog.Error(err.Error())
		return false, fmt.Errorf("failed to send initial request: %w", err)
	}

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

		if err = stream.Send(&pb.UploadFileRequest{
			Content: buf[:n],
		}); err != nil {
			slog.Error(err.Error())
			return false, fmt.Errorf("failed to send file chunk: %w", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		slog.Error(err.Error())
		return false, fmt.Errorf("failed to close and receive response: %w", err)
	}

	return resp.Success, nil
}

func (s *filesService) DownloadFile(ctx context.Context, userID, filePath string, w *io.PipeWriter) error {
	defer w.Close()
	stream, err := s.filesServerClient.DownloadFile(ctx, &pb.DownloadFileRequest{
		UserID:   userID,
		FilePath: filePath,
	})
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to create download stream: %w", err)
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				err = fmt.Errorf("failed to receive download file response: %w", err)
				slog.Error(err.Error())
				return err
			}
		}

		_, err = w.Write(resp.Content)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			err = fmt.Errorf("failed to write download file response: %w", err)
			slog.Error(err.Error())
			return err
		}
	}

	if err = stream.CloseSend(); err != nil {
		err = fmt.Errorf("failed to close send: %w", err)
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (s *filesService) RemoveFile(ctx context.Context, userID, filePath string) (bool, error) {
	resp, err := s.filesServerClient.RemoveFile(ctx, &pb.RemoveFileRequest{
		UserID:   userID,
		FilePath: filePath,
	})
	if err != nil {
		slog.Error(err.Error())
		return false, err
	}
	return resp.Success, nil
}

func NewFilesService(client pb.FileServiceClient) FilesService {
	return &filesService{
		filesServerClient: client,
	}
}
