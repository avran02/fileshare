package dto

import (
	"io"
)

type DownloadStreamResponse struct {
	Success bool
	Content []byte

	r *io.PipeReader
	w *io.PipeWriter
}

func (r *DownloadStreamResponse) Read(buf []byte) (int, error) {
	return r.r.Read(buf)
}

func (r *DownloadStreamResponse) Write(buf []byte) (int, error) {
	return r.w.Write(buf)
}

func (r *DownloadStreamResponse) CloseReader() error {
	return r.r.Close()
}

func (r *DownloadStreamResponse) CloseWriter() error {
	return r.w.Close()
}

func NewDownloadFileStreamResponse() *DownloadStreamResponse {
	pr, pw := io.Pipe()
	return &DownloadStreamResponse{
		r: pr,
		w: pw,
	}
}
