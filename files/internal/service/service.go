package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/avran02/fileshare/files/internal/config"
	"github.com/avran02/fileshare/files/pb"
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
	UploadFile(ctx context.Context, chunk []byte, bucketName, filePath string) error
	DownloadFile(ctx context.Context, bucketName, filePath string) (*minio.Object, error)
	RemoveFile(ctx context.Context, bucketName, filePath string) error
}

type filesService struct {
	minio *minio.Client
}

func (s *filesService) ListFiles(ctx context.Context, bucketName string, dir string) ([]*pb.FileInfo, error) {
	slog.Info("List files in " + dir)
	err := s.checkBucketExists(ctx, bucketName)
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
		slog.Info("Added file: " + object.Key + " to list" + object.LastModified.String())
	}

	return files, nil
}

func (s *filesService) RegisterUser(ctx context.Context, bucketName string) error {
	err := s.checkBucketExists(ctx, bucketName)
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

func (s *filesService) UploadFile(ctx context.Context, chunk []byte, bucketName, filePath string) error {
	err := s.checkBucketExists(ctx, bucketName)
	if err != nil && !errors.Is(err, ErrorBucketExists) {
		return err
	}

	objName := s.getObjectName(bucketName, filePath)

	reader := bytes.NewReader(chunk)

	_, err = s.minio.PutObject(ctx, bucketName, objName, reader, -1, minio.PutObjectOptions{})
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *filesService) DownloadFile(ctx context.Context, bucketName, filePath string) (*minio.Object, error) {
	// TODO: make stream
	err := s.checkBucketExists(ctx, bucketName)
	if err != nil && !errors.Is(err, ErrorBucketExists) {
		return nil, err
	}

	objName := s.getObjectName(bucketName, filePath)

	o, err := s.minio.GetObject(ctx, bucketName, objName, getObjectOptions)
	if err != nil {
		slog.Error(err.Error())
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return o, nil
}

func (s *filesService) RemoveFile(ctx context.Context, bucketName, filePath string) error {
	err := s.checkBucketExists(ctx, bucketName)
	if err != nil && !errors.Is(err, ErrorBucketExists) {
		return err
	}

	return s.minio.RemoveObject(ctx, bucketName, filePath, minio.RemoveObjectOptions{})
}

func (s *filesService) getObjectName(bucketName, filePath string) string {
	return bucketName + "/" + filePath
}

func (s *filesService) checkBucketExists(ctx context.Context, bucketName string) error {
	exists, err := s.minio.BucketExists(ctx, bucketName)
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		return ErrorBucketExists
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
