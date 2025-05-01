package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/bigquery"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/compute"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/model"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/pkg/constants"
	"go.uber.org/zap"
)

type Processor struct {
	traceId       string
	logger        *zap.Logger
	fileUrl       []string
	client        *bigquery.Client
	compute       *compute.Compute
	projectId     string
	projectRegion string
	jobName       string
}

func NewProcessor(traceId string, fileUrl []string, logger *zap.Logger, client *bigquery.Client, compute *compute.Compute, projectId string, region string, jobName string) *Processor {
	return &Processor{
		traceId:       traceId,
		logger:        logger,
		fileUrl:       fileUrl,
		client:        client,
		compute:       compute,
		projectId:     projectId,
		projectRegion: region,
		jobName:       jobName,
	}
}

func (p *Processor) AnalyzeFileUrls(ctx context.Context, fileUrls []string) []model.FileInfo {
	var requests []model.FileInfo
	for _, fileUrl := range fileUrls {
		fileInfo := p.analyzeFile(ctx, fileUrl)
		err := p.decideCompute(ctx, fileInfo)
		if err != nil {
			fileInfo.Error = err.Error()
		}
		requests = append(requests, fileInfo)
	}

	return requests
}

func (p *Processor) decideCompute(ctx context.Context, request model.FileInfo) error {
	ext := request.FileExtension
	switch ext {
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
			FunctionName: constants.APPLICATION_NAME,
		})

		args := []string{request.TraceId, request.FIleUrl, request.FileSizeBytes}
		err := p.compute.TriggerFileStreamerJob(ctx, p.projectId, p.projectRegion, constants.CLOUD_RUN_JOB_NAME, args)
		if err != nil {
			p.logger.Error("error triggering cloud run job",
				zap.String("applicationName", constants.APPLICATION_NAME),
				zap.String("traceId", p.traceId),
				zap.String("fileSize", request.FileSize),
				zap.Error(err))

			p.client.LogAuditData(ctx, model.AuditEvent{
				TraceID:      p.traceId,
				ContractId:   p.traceId,
				Event:        constants.FAILED_TRIGGER_CLOUD_RUN_JOB,
				Status:       constants.FAILED,
				Timestamp:    time.Now(),
				FunctionName: constants.APPLICATION_NAME,
			})

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
			Event:        constants.TRIGGER_CLOUD_BATCH_JOB,
			Status:       constants.IN_PROGRESS,
			Timestamp:    time.Now(),
			FunctionName: constants.APPLICATION_NAME,
		})

		arguments := model.Arguments{
			TraceId:        p.traceId,
			FIleUrl:        request.FIleUrl,
			FileName:       request.FileName,
			RangeSupported: request.RangeSupported,
			FileExtension:  request.FileExtension,
			FileSize:       request.FileSize,
			ContentType:    request.ContentType,
		}

		argsJSON, error := json.Marshal(arguments)
		if error != nil {
			p.logger.Error("failed to marshal arguments to JSON",
				zap.String("applicationName", constants.APPLICATION_NAME),
				zap.String("traceId", p.traceId),
				zap.Any("arguments", arguments),
				zap.Error(error))
			return fmt.Errorf("failed to marshal arguments: %w", error)
		}

		event := model.ContractFileEvent{
			TraceID:      request.TraceId,
			ContractID:   request.TraceId,
			Status:       constants.STARTED,
			Timestamp:    time.Now(),
			FunctionName: constants.APPLICATION_NAME,
			Arguments:    string(argsJSON),
			Environment:  constants.ENVIRONMENT,
		}

		err := p.client.ContractFileQueue(ctx, event)
		if err != nil {
			p.logger.Error("error inserting data into `contract_file_queue` table",
				zap.String("applicationName", constants.APPLICATION_NAME),
				zap.String("traceId", p.traceId),
				zap.String("fileSize", request.FileSize),
				zap.Error(err))

			p.client.LogAuditData(ctx, model.AuditEvent{
				TraceID:      p.traceId,
				ContractId:   p.traceId,
				Event:        constants.FAILED_TRIGGER_CLOUD_BATCH_JOB,
				Status:       constants.FAILED,
				Timestamp:    time.Now(),
				FunctionName: constants.APPLICATION_NAME,
			})
			return err

		}

		args := []string{request.TraceId, request.FIleUrl, request.FileName}
		err = p.compute.TriggerFileStreamerJob(ctx, p.projectId, p.projectRegion, constants.ZIP_DOWNLOADER_JOB_NAME, args)
		if err != nil {
			p.logger.Error("error triggering cloud run job",
				zap.String("applicationName", constants.APPLICATION_NAME),
				zap.String("traceId", p.traceId),
				zap.String("fileSize", request.FileSize),
				zap.Error(err))

			p.client.LogAuditData(ctx, model.AuditEvent{
				TraceID:      p.traceId,
				ContractId:   p.traceId,
				Event:        constants.FAILED_TRIGGER_CLOUD_RUN_JOB,
				Status:       constants.FAILED,
				Timestamp:    time.Now(),
				FunctionName: constants.APPLICATION_NAME,
			})

			return err
		}

		return nil

	default:
		return nil
	}
}

func (p *Processor) getFileNameFromURL(rawURL string) (string, error) {
	parsedUrl, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(parsedUrl.Path), nil
}

func (p *Processor) analyzeFile(ctx context.Context, fileUrl string) model.FileInfo {
	var info model.FileInfo

	parsedUrl, err := url.Parse(fileUrl)
	if err != nil {
		p.logger.Error("Invalid URL",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.Error(err))
		info.Error = fmt.Sprintf("Invalid URL %s: %v", fileUrl, err)
		return info
	}

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

	p.logger.Info("Process completed",
		zap.String("applicationName", constants.APPLICATION_NAME),
		zap.String("traceId", p.traceId),
		zap.String("extension", info.FileExtension), zap.String("file size", info.FileSize),
		zap.String("content type", info.ContentType))

	return info
}
