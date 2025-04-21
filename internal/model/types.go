package model

import "time"

type FileInfo struct {
	TraceId       string  `json:"traceid"`
	FileExtension string  `json:"fileExtenstion,omitempty"`
	FileSize      string  `json:"fileSize,omitempty"`
	FileSizeFloat float64 `json:"-"`
	ContentType   string  `json:"contentType,omitempty"`
	Error         string  `json:"error,omitempty"`
}

type AuditEvent struct {
	TraceID      string    `bigquery:"traceid"`
	ContractId   string    `json:"contractId"`
	Event        string    `bigquery:"event"`
	Status       string    `bigquery:"status"`
	Timestamp    time.Time `bigquery:"timestamp"`
	FunctionName string    `bigquery:"functionName"`
	Environment  string    `bigquery:"environment"`
}

type RequestBody struct {
	FileUrl     []string `json:"fileUrl"`
	RequestUUID string   `json:"requestUUID"`
}
