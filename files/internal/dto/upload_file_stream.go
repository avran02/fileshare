package dto

import (
	"errors"
	"io"
)

type UploadFileStreamRequest struct {
	UserID   string
	FilePath string
	reader   *io.PipeReader
	writer   *io.PipeWriter

	Content []byte
}

func (r *UploadFileStreamRequest) GetData(userID, filePath string) error {
	if userID == "" {
		return errors.New("empty user id")
	}

	r.UserID = userID

	if filePath == "" {
		return errors.New("empty file path")
	}

	r.FilePath = filePath

	pr, pw := io.Pipe()
	r.reader = pr
	r.writer = pw

	return nil
}

func (r *UploadFileStreamRequest) Read(buf []byte) (int, error) {
	return r.reader.Read(buf)
}

func (r *UploadFileStreamRequest) Write(buf []byte) (int, error) {
	return (*r.writer).Write(buf)
}

func (r *UploadFileStreamRequest) CloseWriter() error {
	return r.writer.Close()
}

func (r *UploadFileStreamRequest) CloseReader() error {
	return r.reader.Close()
}
