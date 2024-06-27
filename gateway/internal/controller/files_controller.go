package controller

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/avran02/fileshare/gateway/internal/dto"
	"github.com/avran02/fileshare/gateway/internal/service"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const defaultChunkSize = 1024

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
	ctx := r.Context()
	userID := r.URL.Query().Get("userID")
	filePath := r.URL.Query().Get("filePath")

	if userID == "" || filePath == "" {
		slog.Error("userID or filePath is empty")
		http.Error(w, "userID or filePath is empty", http.StatusBadRequest)
		return
	}

	slog.Info("userID: " + userID + " filePath: " + filePath)
	resp := dto.NewDownloadFileResponse()
	go func() {

		err := c.service.DownloadFile(ctx, userID, filePath, *resp)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.CloseReader()
		slog.Info("got resp in controller.Download")

		chunk := make([]byte, defaultChunkSize)
		_, err = resp.Read(chunk)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		slog.Info("read chunk with len: " + fmt.Sprint(len(chunk)))
	}()

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err := io.Copy(w, resp)
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

	ok, err := c.service.UploadFile(r.Context(), req.File, req.UserID, req.FilePath)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(fmt.Sprint(ok)))
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
