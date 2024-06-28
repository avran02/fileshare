package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"

	"github.com/avran02/fileshare/files/internal/config"
	"github.com/avran02/fileshare/files/internal/dto"
	pb "github.com/avran02/fileshare/proto/filespb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	getObjectOptions  = minio.GetObjectOptions{}
	makeBucketOptions = minio.MakeBucketOptions{
		Region: config.DefaultLocation,
	}
)

type FilesService interface {
	RegisterUser(ctx context.Context, bucketName string) error
	ListFiles(ctx context.Context, bucketName, dir string) ([]*pb.FileInfo, error)
	UploadFile(ctx context.Context, req *dto.UploadFileStreamRequest) error
	DownloadFile(ctx context.Context, bucketName, filePath string) (io.ReadCloser, error)
	RemoveFile(ctx context.Context, bucketName, filePath string) error
}

type filesService struct {
	minio *minio.Client
}

func (s *filesService) ListFiles(ctx context.Context, bucketName string, dir string) ([]*pb.FileInfo, error) {
	slog.Info("List files in " + dir)
	err := s.createBucketIfNotExists(ctx, bucketName)
	if err != nil && !errors.Is(err, ErrorBucketExists) {
		return nil, err
	}

	objChan := s.minio.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: false,
	})

	files := make([]*pb.FileInfo, 0, len(objChan))

	for object := range objChan {
		if object.Err != nil {
			slog.Error(object.Err.Error())
			return nil, fmt.Errorf("failed to list objects:\n%w", object.Err)
		}

		files = append(files, &pb.FileInfo{
			Name:         object.Key,
			Size:         object.Size,
			LastModified: timestamppb.New(object.LastModified),
		})
	}

	return files, nil
}

func (s *filesService) RegisterUser(ctx context.Context, bucketName string) error {
	err := s.createBucketIfNotExists(ctx, bucketName)
	if err != nil && !errors.Is(err, ErrorBucketExists) {
		return err
	}

	err = s.minio.MakeBucket(ctx, bucketName, makeBucketOptions)
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

func (s *filesService) UploadFile(ctx context.Context, req *dto.UploadFileStreamRequest) error {
	if err := s.createBucketIfNotExists(ctx, req.UserID); err != nil {
		return err
	}

	_, err := s.minio.PutObject(ctx, req.UserID, req.FilePath, req, -1, minio.PutObjectOptions{})
	if err != nil {
		if errors.Is(err, io.EOF) {
			slog.Info("closed in EOF block")
			return nil
		}
		slog.Error(err.Error())
		return fmt.Errorf("failed to upload file: %w", err)
	}

	slog.Info("Uploaded file: " + req.FilePath)

	return nil
}

func (s *filesService) DownloadFile(ctx context.Context, bucketName, filePath string) (io.ReadCloser, error) {
	if err := s.createBucketIfNotExists(ctx, bucketName); err != nil {
		return nil, err
	}

	o, err := s.minio.GetObject(ctx, bucketName, filePath, getObjectOptions)
	if err != nil {
		err = fmt.Errorf("failed to get object: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return o, nil
}

func (s *filesService) RemoveFile(ctx context.Context, bucketName, filePath string) error {
	if err := s.createBucketIfNotExists(ctx, bucketName); err != nil {
		return err
	}

	return s.minio.RemoveObject(ctx, bucketName, filePath, minio.RemoveObjectOptions{})
}

func (s *filesService) createBucketIfNotExists(ctx context.Context, bucketName string) error {
	exists, err := s.minio.BucketExists(ctx, bucketName)
	if err != nil {
		err = fmt.Errorf("failed to check if bucket exists: %w", err)
		slog.Error(err.Error())
		return err
	}

	if !exists {
		err = s.minio.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: config.DefaultLocation})
		if err != nil {
			err = fmt.Errorf("failed to create bucket: %w", err)
			slog.Error(err.Error())
			return err
		}
	}

	return nil
}

func New(conf config.Minio) FilesService {
	minioClient, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Region: config.DefaultLocation,
		Secure: false,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	return &filesService{
		minio: minioClient,
	}
}
