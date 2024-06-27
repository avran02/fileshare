package dto

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
)

const maxFileSize = 1024 * 1024 * 1024

type FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
}

type ListFilesResponse struct {
	Files []FileInfo `json:"files"`
}

type UploadFileRequest struct {
	UserID   string         `json:"userID"`
	FilePath string         `json:"filePath"`
	File     multipart.File `json:"file"`
}

func NewUploadFileRequestFromHTTPForm(req *http.Request) (*UploadFileRequest, error) {
	if err := req.ParseMultipartForm(maxFileSize); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return &UploadFileRequest{
		FilePath: req.FormValue("filePath"),
		UserID:   req.FormValue("userID"),
		File:     file,
	}, nil
}

type UploadFileResponse struct {
	Success bool `json:"success"`
}

type DownloadFileRequest struct {
	UserID   string `json:"userID"`
	FilePath string `json:"filePath"`
}

type RemoveFileRequest struct {
	UserID   string `json:"userID"`
	FilePath string `json:"filePath"`
}

type RemoveFileResponse struct {
	Success bool `json:"success"`
}
