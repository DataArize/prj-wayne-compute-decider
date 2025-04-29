package processor

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/bigquery"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/model"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/pkg/constants"
	"go.uber.org/zap"
)

type Processor struct {
	traceId string
	logger  *zap.Logger
	fileUrl []string
	client  *bigquery.Client
}

func NewProcessor(traceId string, fileUrl []string, logger *zap.Logger, client *bigquery.Client) *Processor {
	return &Processor{
		traceId: traceId,
		logger:  logger,
		fileUrl: fileUrl,
		client:  client,
	}
}

func (p *Processor) AnalyzeFileUrls(ctx context.Context, fileUrls []string) []model.FileInfo {
	var requests []model.FileInfo
	for _, fileUrl := range fileUrls {
		fileInfo := p.analyzeFile(ctx, fileUrl)
		p.decideCompute(fileInfo)
		requests = append(requests, fileInfo)
	}

	return requests
}

func (p *Processor) decideCompute(request model.FileInfo) error {
	ext := request.FileExtension
	estimatedFileSize := 4 * request.FileSizeFloat
	switch {
	case ext == constants.GZ:
		p.logger.Info("File extension is GZ",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.String("fileSize", request.FileSize))
		return nil
	case estimatedFileSize > constants.MAX_FILE_SIZE:
		p.logger.Info("file size is greater than max file size for cloud run",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", p.traceId),
			zap.String("fileName", request.FileSize),
			zap.Int("maxFileSize", constants.MAX_FILE_SIZE))

	default:
		return nil
	}
	return nil
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
	info.FileSize = fmt.Sprintf("%.2f GB", fileSizeGB)
	info.ContentType = resp.Header.Get(constants.CONTENT_TYPE)
	info.TraceId = p.traceId

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
