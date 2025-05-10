// Package processor provides logic to analyze files and determine whether to trigger
// compute jobs based on file metadata such as size and extension.
package processor

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/bigquery"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/compute"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/gcs"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/model"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/pkg/constants"
	"go.uber.org/zap"
)

// Processor coordinates the logic for analyzing files and deciding compute actions.
type Processor struct {
	traceId       string
	logger        *zap.Logger
	fileUrl       []string
	client        *bigquery.Client
	gcs           *gcs.GCSClient
	compute       *compute.Compute
	projectId     string
	projectRegion string
	jobName       string
}

// NewProcessor creates and returns a new instance of Processor with all required dependencies.
func NewProcessor(traceId string, fileUrl []string, logger *zap.Logger, client *bigquery.Client, compute *compute.Compute, projectId string, region string, jobName string, gcs *gcs.GCSClient) *Processor {
	return &Processor{
		traceId:       traceId,
		logger:        logger,
		fileUrl:       fileUrl,
		client:        client,
		compute:       compute,
		projectId:     projectId,
		projectRegion: region,
		jobName:       jobName,
		gcs:           gcs,
	}
}

// AnalyzeFileUrls iterates over the provided file URLs, analyzes each one,
// and determines whether to trigger a compute job.
func (p *Processor) AnalyzeFileUrls(ctx context.Context, fileUrls []string, requestUUID string) []model.FileInfo {
	var requests []model.FileInfo
	for _, fileUrl := range fileUrls {
		fileInfo := p.analyzeFile(ctx, fileUrl)
		isProcessed, err := p.gcs.CheckAlreadyProcessed(fileInfo, ctx, requestUUID)
		if err != nil {
			p.client.LogAuditData(ctx, model.AuditEvent{
				TraceID:      p.traceId,
				ContractId:   p.traceId,
				Event:        constants.FAILED_TO_CHECK_IF_FILE_EXISTS,
				FileUrl:      fileUrl,
				Status:       constants.FAILED,
				Timestamp:    time.Now(),
				FunctionName: constants.APPLICATION_NAME,
			})
			fileInfo.Error = err.Error()
		} else if !isProcessed {
			err := p.decideCompute(ctx, fileInfo)
			if err != nil {
				p.client.LogAuditData(ctx, model.AuditEvent{
					TraceID:      p.traceId,
					ContractId:   p.traceId,
					Event:        constants.FAILED_TRIGGER_CLOUD_RUN_JOB,
					FileUrl:      fileUrl,
					Status:       constants.FAILED,
					Timestamp:    time.Now(),
					FunctionName: constants.APPLICATION_NAME,
				})
				fileInfo.Error = err.Error()
			}

		}
		requests = append(requests, fileInfo)
	}

	return requests
}

// decideCompute determines the compute action to take based on file extension (e.g. .gz or .zip).
// It logs appropriate audit events and triggers cloud run jobs or queues messages.
func (p *Processor) decideCompute(ctx context.Context, request model.FileInfo) error {
	ext := request.FileExtension
	switch ext {
	case constants.JSON:
		p.logger.Info("File extension is JSON",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.String("fileSize", request.FileSize))

		p.client.LogAuditData(ctx, model.AuditEvent{
			TraceID:      p.traceId,
			ContractId:   p.traceId,
			Event:        constants.TRIGGER_CLOUD_RUN_JOB,
			Status:       constants.IN_PROGRESS,
			Timestamp:    time.Now(),
			FileUrl:      request.FIleUrl,
			FunctionName: constants.APPLICATION_NAME,
		})

		args := []string{request.TraceId, request.FIleUrl, request.FileSizeBytes, request.RequestUUID}
		err := p.compute.TriggerFileStreamerJob(ctx, p.projectId, p.projectRegion, constants.CLOUD_RUN_JOB_NAME, args)
		if err != nil {
			p.logger.Error("error triggering cloud run job",
				zap.String("applicationName", constants.APPLICATION_NAME),
				zap.String("traceId", p.traceId),
				zap.String("fileSize", request.FileSize),
				zap.Error(err))
			return err
		}
		return nil

	case constants.GZ:
		p.logger.Info("File extension is GZ",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.String("fileSize", request.FileSize))

		p.client.LogAuditData(ctx, model.AuditEvent{
			TraceID:      p.traceId,
			ContractId:   p.traceId,
			Event:        constants.TRIGGER_CLOUD_RUN_JOB,
			Status:       constants.IN_PROGRESS,
			Timestamp:    time.Now(),
			FileUrl:      request.FIleUrl,
			FunctionName: constants.APPLICATION_NAME,
		})

		args := []string{request.TraceId, request.FIleUrl, request.FileSizeBytes, request.RequestUUID}
		err := p.compute.TriggerFileStreamerJob(ctx, p.projectId, p.projectRegion, constants.GZ_JOB_NAME, args)
		if err != nil {
			p.logger.Error("error triggering cloud run job",
				zap.String("applicationName", constants.APPLICATION_NAME),
				zap.String("traceId", p.traceId),
				zap.String("fileSize", request.FileSize),
				zap.Error(err))
			return err
		}
		return nil
	case constants.ZIP:
		p.logger.Info("File extension is ZIP",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.String("fileSize", request.FileSize))

		p.client.LogAuditData(ctx, model.AuditEvent{
			TraceID:      p.traceId,
			ContractId:   p.traceId,
			Event:        constants.TRIGGER_CLOUD_RUN_JOB,
			Status:       constants.IN_PROGRESS,
			Timestamp:    time.Now(),
			FileUrl:      request.FIleUrl,
			FunctionName: constants.APPLICATION_NAME,
		})

		args := []string{request.TraceId, request.FIleUrl, request.FileName}
		err := p.compute.TriggerFileStreamerJob(ctx, p.projectId, p.projectRegion, constants.ZIP_DOWNLOADER_JOB_NAME, args)
		if err != nil {
			p.logger.Error("error triggering cloud run job",
				zap.String("applicationName", constants.APPLICATION_NAME),
				zap.String("traceId", p.traceId),
				zap.String("fileSize", request.FileSize),
				zap.Error(err))

			return err
		}
		return nil

	default:
		return nil
	}
}

// getFileNameFromURL extracts the file name from a URL path.
func (p *Processor) getFileNameFromURL(rawURL string) (string, error) {
	parsedUrl, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(parsedUrl.Path), nil
}

// analyzeFile performs a HEAD request to gather metadata about the file,
// such as content length, content type, extension, and range support.
func (p *Processor) analyzeFile(ctx context.Context, fileUrl string, requestUUID string) model.FileInfo {
	var info model.FileInfo
	info.RequestUUID = requestUUID

	parsedUrl, err := url.Parse(fileUrl)
	if err != nil {
		p.logger.Error("Invalid URL",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.Error(err))
		info.Error = fmt.Sprintf("Invalid URL %s: %v", fileUrl, err)
		return info
	}
	p.client.LogAuditData(ctx, model.AuditEvent{
		TraceID:      p.traceId,
		ContractId:   p.traceId,
		Event:        constants.ANALYZE_FILE_STARTED,
		Status:       constants.IN_PROGRESS,
		Timestamp:    time.Now(),
		FileUrl:      fileUrl,
		FunctionName: constants.APPLICATION_NAME,
	})

	info.FileExtension = path.Ext(parsedUrl.Path)

	req, err := http.NewRequestWithContext(ctx, constants.HEAD, fileUrl, nil)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to create HEAD request URL: %s, error : %v", fileUrl, err)
		p.logger.Error("unable to create HEAD Request",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.Error(err))
		return info
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to execute HEAD request URL: %s, error : %v", fileUrl, err)
		p.logger.Error("unable to create HEAD Request",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.Error(err))
		return info
	}

	defer resp.Body.Close()

	fileSize := resp.Header.Get(constants.CONTENT_LENGTH)
	acceptRanges := resp.Header.Get(constants.RANGE_SUPPORTED)
	if acceptRanges == constants.BYTES {
		info.RangeSupported = true
	} else {
		info.RangeSupported = false
	}

	fileName, err := p.getFileNameFromURL(fileUrl)
	if err != nil {
		info.Error = fmt.Sprintf("error extracting file name, fileURL: %s, error: %v", fileUrl, err)
		p.logger.Error("error extracting file name",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.Error(err))
		return info
	}

	info.FileName = fileName

	fileSizeConverted, err := strconv.ParseInt(fileSize, 10, 64)
	if err != nil {
		info.Error = fmt.Sprintf("error parsing content length, fileURL: %s, error: %v", fileUrl, err)
		p.logger.Error("error parsing content length",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.Error(err))
		return info
	}

	fileSizeGB := float64(fileSizeConverted) / constants.FILE_SIZE_BYTES
	info.FileSizeFloat = fileSizeGB
	info.FileSizeBytes = fileSize
	info.FileSize = fmt.Sprintf("%.2f GB", fileSizeGB)
	info.ContentType = resp.Header.Get(constants.CONTENT_TYPE)
	info.TraceId = p.traceId
	info.FIleUrl = fileUrl

	if info.FileExtension == "" && info.ContentType != "" {
		parts := strings.Split(info.ContentType, "/")
		if len(parts) == 2 {
			info.FileExtension = "." + parts[1]
		}
	}

	p.client.LogAuditData(ctx, model.AuditEvent{
		TraceID:      p.traceId,
		ContractId:   p.traceId,
		Event:        constants.ANALYZE_FILE_COMPLETED,
		Status:       constants.IN_PROGRESS,
		Timestamp:    time.Now(),
		FileUrl:      fileUrl,
		FunctionName: constants.APPLICATION_NAME,
	})

	p.logger.Info("Process completed",
		zap.String("applicationName", constants.APPLICATION_NAME),
		zap.String("traceId", p.traceId),
		zap.String("extension", info.FileExtension), zap.String("file size", info.FileSize),
		zap.String("content type", info.ContentType))

	return info
}
