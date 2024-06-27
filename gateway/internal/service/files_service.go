package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/avran02/fileshare/files/pb"
	"github.com/avran02/fileshare/gateway/internal/dto"
)

var chankSize = 1024 * 1024

type FilesService interface {
	ListFiles(ctx context.Context, userID, filePath string) ([]*pb.FileInfo, error)
	UploadFile(ctx context.Context, reader io.Reader, userID, filePath string) (bool, error)
	DownloadFile(ctx context.Context, userID, filePath string, resp dto.DownloadFileResponse) error
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

func (s *filesService) DownloadFile(ctx context.Context, userID, filePath string, r dto.DownloadFileResponse) error {
	streamErrChan := make(chan error, 1)

	stream, err := s.filesServerClient.DownloadFile(ctx, &pb.DownloadFileRequest{
		UserID:   userID,
		FilePath: filePath,
	})
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to create download stream: %w", err)
	}

	go func() {
		defer r.CloseWriter()
		defer close(streamErrChan)

		for {
			slog.Info("requesting bytes")
			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					slog.Info("End of file")
					break
				}
				slog.Error(err.Error())
				streamErrChan <- fmt.Errorf("failed to receive download file response: %w", err)
				return
			}

			if len(resp.Content) == 0 {
				slog.Info("No content")
				if resp.Success {
					slog.Info("Success")
					break
				}
				slog.Warn("No content and no success")
				break
			}

			slog.Info("Got " + fmt.Sprint(len(resp.Content)) + " bytes")

			n, err := r.Write(resp.Content)
			slog.Info("n: " + fmt.Sprint(n) + " bytes")
			if err != nil {
				slog.Error("Read error: " + err.Error())
				if errors.Is(err, io.EOF) && n == 0 {
					slog.Info("EOF")
					break
				} else {
					slog.Error(err.Error())
					streamErrChan <- fmt.Errorf("failed to read download file response: %w", err)
				}
			}
			slog.Info("Received " + fmt.Sprint(n) + " bytes")
		}

		err = stream.CloseSend()
		if err != nil {
			slog.Error(err.Error())
			streamErrChan <- fmt.Errorf("failed to close send: %w", err)
		}
	}()

	err = <-streamErrChan
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
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
