package dto

import (
	"errors"
	"io"
)

var (
	ErrEmptyUserID   = errors.New("empty user id")
	ErrEmptyFilePath = errors.New("empty file path")
)

type UploadFileStreamRequest struct {
	UserID   string
	FilePath string
	reader   *io.PipeReader
	writer   *io.PipeWriter

	Content []byte
}

func (r *UploadFileStreamRequest) Read(buf []byte) (int, error) {
	return r.reader.Read(buf)
}

func (r *UploadFileStreamRequest) Write(buf []byte) (int, error) {
	return (*r.writer).Write(buf)
}

func (r *UploadFileStreamRequest) CloseWriter() {
	r.writer.Close()
}

func (r *UploadFileStreamRequest) CloseReader() {
	r.reader.Close()
}

func NewUploadFileStreamRequest(userID, filePath string) (*UploadFileStreamRequest, error) {
	if userID == "" {
		return nil, ErrEmptyUserID
	}

	if filePath == "" {
		return nil, ErrEmptyFilePath
	}

	pr, pw := io.Pipe()

	return &UploadFileStreamRequest{
		UserID:   userID,
		FilePath: filePath,

		reader: pr,
		writer: pw,
	}, nil
}
