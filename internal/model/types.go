package model

import "time"

type FileInfo struct {
	TraceId        string  `json:"traceid"`
	RequestUUID    string  `json:"requestUUID"`
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

type Arguments struct {
	TraceId        string `bigquery:"traceid"`
	FIleUrl        string `bigquery:"fileUrl"`
	FileName       string `bigquery:"fileName"`
	RangeSupported bool   `bigquery:"rangeSupported"`
	FileExtension  string `bigquery:"fileExtenstion"`
	FileSize       string `bigquery:"fileSize"`
	ContentType    string `bigquery:"contentType"`
}

type AuditEvent struct {
	TraceID      string    `bigquery:"traceid"`
	ContractId   string    `json:"contractId"`
	Event        string    `bigquery:"event"`
	Status       string    `bigquery:"status"`
	Timestamp    time.Time `bigquery:"createdTimestamp"`
	FunctionName string    `bigquery:"functionName"`
	Environment  string    `bigquery:"environment"`
	Message      string    `bigquery:"message"`
	FileUrl      string    `bigquery:"fileUrl"`
}

type RequestBody struct {
	FileUrl          []string `json:"fileUrl"`
	RequestUUID      string   `json:"requestUUID"`
	ForceProcessFlag bool     `json:"forceProcessFlag"`
}

type ContractFileEvent struct {
	TraceID      string    `bigquery:"traceId"`
	ContractID   string    `bigquery:"contractId"`
	Status       string    `bigquery:"status"` // can be null
	Timestamp    time.Time `bigquery:"timestamp"`
	FunctionName string    `bigquery:"functionName"` // can be null
	Arguments    string    `bigquery:"arguments"`    // nested struct
	Environment  string    `bigquery:"environment"`  // can be null
}
