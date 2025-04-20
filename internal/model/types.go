package model

import "time"

type FileInfo struct {
	TraceId       string `json:"traceid"`
	FileExtension string `json:"fileExtenstion,omitempty"`
	FileSize      string `json:"fileSize,omitempty"`
	ContentType   string `json:"contentType,omitempty"`
	Error         string `json:"error,omitempty"`
}

type AuditEvent struct {
	TraceID   string    `bigquery:"traceid"`
	Event     string    `bigquery:"event"`
	Status    string    `bigquery:"status"`
	Timestamp time.Time `bgiquery:"timestamp"`
}

type RequestBody struct {
	FileUrl     []string `json:"fileUrl"`
	RequestUUID string   `json:"requestUUID"`
}
