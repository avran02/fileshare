package controller

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/avran02/fileshare/gateway/internal/dto"
	"github.com/avran02/fileshare/gateway/internal/middlaware"
	"github.com/avran02/fileshare/gateway/internal/service"
)

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
	userID := ctx.Value(middlaware.ContextUserIDKey).(string)

	pr, pw := io.Pipe()
	streamErrChan := make(chan error, 1)

	filePath := r.URL.Query().Get("filePath")

	filePathParts := strings.Split(filePath, "/")
	fileName := filePathParts[len(filePathParts)-1]

	if userID == "" || filePath == "" {
		slog.Error("userID or filePath is empty")
		http.Error(w, "userID or filePath is empty", http.StatusBadRequest)
		return
	}

	go c.asyncDownloadFileFromGrpcStream(ctx, userID, filePath, pw, streamErrChan)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	_, err := io.Copy(w, pr)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = <-streamErrChan
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *filesController) Upload(w http.ResponseWriter, r *http.Request) {
	slog.Info("Upload a file")

	ctx := r.Context()
	userID := ctx.Value(middlaware.ContextUserIDKey).(string)

	req, err := dto.NewUploadFileRequestFromHTTPForm(r)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = c.service.UploadFile(ctx, req.File, userID, req.FilePath)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("ok"))
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *filesController) Rm(w http.ResponseWriter, r *http.Request) {
	slog.Info("Remove a file")
	ctx := r.Context()
	userID := ctx.Value(middlaware.ContextUserIDKey).(string)

	rawBody := make([]byte, r.ContentLength)
	_, err := r.Body.Read(rawBody)
	if err != nil && err != io.EOF {
		err = fmt.Errorf("failed to read request body: %w", err)
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var body dto.RemoveFileRequest
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal request body: %w", err)
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ok, err := c.service.RemoveFile(ctx, userID, body.FilePath)
	if err != nil {
		err = fmt.Errorf("failed to remove file: %w", err)
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rawResp, err := json.Marshal(dto.RemoveFileResponse{Success: ok})
	if err != nil {
		err = fmt.Errorf("failed to marshal response: %w", err)
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(rawResp)
	if err != nil {
		err = fmt.Errorf("failed to write response: %w", err)
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *filesController) Ls(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(middlaware.ContextUserIDKey).(string)
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

func (c *filesController) asyncDownloadFileFromGrpcStream(ctx context.Context, userID, filePath string, w *io.PipeWriter, streamErrChan chan error) {
	defer close(streamErrChan)

	if err := c.service.DownloadFile(ctx, userID, filePath, w); err != nil {
		slog.Error(err.Error())
		streamErrChan <- fmt.Errorf("failed to create download stream: %w", err)
		return
	}
}

func NewFilesController(service service.FilesService) FilesController {
	return &filesController{
		service: service,
	}
}
