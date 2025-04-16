package model

type FileInfo struct {
	TraceId       string `json:"traceid"`
	FileExtension string `json:"fileExtenstion,omitempty"`
	FileSize      string `json:"fileSize,omitempty"`
	ContentType   string `json:"contentType,omitempty"`
	Error         string `json:"error,omitempty"`
}
