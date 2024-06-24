package dto

import (
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

func (r *UploadFileRequest) FromHTTP(req *http.Request) error {
	err := req.ParseMultipartForm(maxFileSize)
	if err != nil {
		return err
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		return err
	}

	r.UserID = req.FormValue("userID")
	r.FilePath = req.FormValue("filePath")
	r.File = file
	return nil
}

type UploadFileResponse struct {
	Success bool `json:"success"`
}

type DownloadFileRequest struct {
	UserID   string `json:"userID"`
	FilePath string `json:"filePath"`
}

type DownloadFileResponse struct {
	Success bool   `json:"success"`
	Content []byte `json:"content"`
}

type RemoveFileRequest struct {
	UserID   string `json:"userID"`
	FilePath string `json:"filePath"`
}

type RemoveFileResponse struct {
	Success bool `json:"success"`
}
