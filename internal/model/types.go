package model

import "time"

type FileInfo struct {
	TraceId        string  `json:"traceid"`
	FIleUrl        string  `json:"fileUrl"`
	FileName       string  `json:"fileName"`
	RangeSupported bool    `json:"rangeSupported"`
	FileExtension  string  `json:"fileExtenstion,omitempty"`
	FileSize       string  `json:"fileSize,omitempty"`
	FileSizeFloat  float64 `json:"-"`
	FileSizeBytes  string  `json:"-"`
	ContentType    string  `json:"contentType,omitempty"`
	Error          string  `json:"error,omitempty"`
}

type AuditEvent struct {
	TraceID      string    `bigquery:"traceid"`
	ContractId   string    `json:"contractId"`
	Event        string    `bigquery:"event"`
	Status       string    `bigquery:"status"`
	Timestamp    time.Time `bigquery:"timestamp"`
	FunctionName string    `bigquery:"functionName"`
	Environment  string    `bigquery:"environment"`
	Message      string    `bigquery:"message"`
}

type RequestBody struct {
	FileUrl     []string `json:"fileUrl"`
	RequestUUID string   `json:"requestUUID"`
}

type ContractFileEvent struct {
	TraceID      string    `bigquery:"traceId"`
	ContractID   string    `bigquery:"contractId"`
	Status       string    `bigquery:"status"` // can be null
	Timestamp    time.Time `bigquery:"timestamp"`
	FunctionName string    `bigquery:"functionName"` // can be null
	Arguments    FileInfo  `bigquery:"arguments"`    // nested struct
	Environment  string    `bigquery:"environment"`  // can be null
}
