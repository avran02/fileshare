package controller

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/avran02/fileshare/gateway/internal/dto"
	"github.com/avran02/fileshare/gateway/internal/service"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type FilesController interface {
	Download(w http.ResponseWriter, r *http.Request)
	Upload(w http.ResponseWriter, r *http.Request)
	Rm(w http.ResponseWriter, r *http.Request)
	Ls(w http.ResponseWriter, r *http.Request)
}

type filesController struct {
	service service.FilesService
}

func (c *filesController) Download(w http.ResponseWriter, r *http.Request) {
	slog.Info("Download a file")

	// TODO: add gRPC call
	c.service.DownloadFile()
	_, err := w.Write([]byte("Download a file"))
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func (c *filesController) Upload(w http.ResponseWriter, r *http.Request) {
	slog.Info("Upload a file")

	req := dto.UploadFileRequest{}
	err := req.FromHTTP(r)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok, err := c.service.UploadFile(r.Context(), r.Body, req.UserID, req.FilePath)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("Upload a file"))
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func (c *filesController) Rm(w http.ResponseWriter, r *http.Request) {
	slog.Info("Remove a file")

	rawBody := make([]byte, r.ContentLength)
	_, err := r.Body.Read(rawBody)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var body dto.RemoveFileRequest
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ok, err := c.service.RemoveFile(body.UserID, body.FilePath)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rawResp, err := json.Marshal(dto.RemoveFileResponse{Success: ok})
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(rawResp)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *filesController) Ls(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.URL.Query().Get("userID")
	filePath := r.URL.Query().Get("filePath")
	files, err := c.service.ListFiles(ctx, userID, filePath)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	if files == nil {
		slog.Warn("files is nil")
		return
	}

	slog.Info("Found " + fmt.Sprint(len(files)) + " files")

	respFiles := make([]dto.FileInfo, len(files))
	for i, file := range files {
		fmt.Println(file)
		if file == nil {
			slog.Warn("file is nil")
			return
		}
		respFiles[i] = dto.FileInfo{
			Name:         file.Name,
			Size:         file.Size,
			LastModified: file.LastModified.AsTime(),
		}
	}
	err = json.NewEncoder(w).Encode(dto.ListFilesResponse{Files: respFiles})
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func NewFilesController(service service.FilesService) FilesController {
	return &filesController{
		service: service,
	}
}