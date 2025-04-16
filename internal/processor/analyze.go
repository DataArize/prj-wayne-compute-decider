package processor

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/model"
	"go.uber.org/zap"
)

type Processor struct {
	logger  *zap.Logger
	fileUrl string
}

const (
	CONTENT_TYPE     = "Content-Type"
	APPLICATION_JSON = "application/json"
	HEAD             = "HEAD"
	CONTENT_LENGTH   = "Content-Length"
)

func NewProcessor(fileUrl string, logger *zap.Logger) *Processor {
	return &Processor{
		logger:  logger,
		fileUrl: fileUrl,
	}
}

func (p *Processor) AnalyzeFile(ctx context.Context, fileUrl string, logger *zap.Logger) model.FileInfo {
	var info model.FileInfo

	parsedUrl, err := url.Parse(fileUrl)
	if err != nil {
		logger.Error("Invalid URL", zap.Error(err))
		info.Error = fmt.Sprintf("Invalid URL : %v", err)
		return info
	}

	info.FileExtension = path.Ext(parsedUrl.Path)

	req, err := http.NewRequestWithContext(ctx, HEAD, fileUrl, nil)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to create HEAD request: %v", err)
		logger.Error("unable to create HEAD Request", zap.Error(err))
		return info
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to execute HEAD request: %v", err)
		logger.Error("unable to create HEAD Request", zap.Error(err))
		return info
	}

	defer resp.Body.Close()

	info.FileSize = resp.Header.Get(CONTENT_LENGTH)
	info.ContentType = resp.Header.Get(CONTENT_TYPE)

	if info.FileExtension == "" && info.ContentType != "" {
		parts := strings.Split(info.ContentType, "/")
		if len(parts) == 2 {
			info.FileExtension = "." + parts[1]
		}
	}

	logger.Info("Process completed", zap.String("extension", info.FileExtension), zap.String("file size", info.FileSize),
		zap.String("content type", info.ContentType))

	return info
}
