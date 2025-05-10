package gcs

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/model"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/pkg/constants"
	"go.uber.org/zap"

	"cloud.google.com/go/storage"
)

type GCSClient struct {
	logger     *zap.Logger
	traceId    string
	bucketName string
	gcsClient  *storage.Client
}

func NewGCSClient(logger *zap.Logger, bucketName string, traceId string, ctx context.Context) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		logger.Error("unable to create storage client",
			zap.String("ApplicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))
		return nil, err
	}

	c := &GCSClient{
		logger:     logger,
		bucketName: bucketName,
		gcsClient:  client,
		traceId:    traceId,
	}

	return c, nil
}

func (c *GCSClient) CheckAlreadyProcessed(fileInfo model.FileInfo, ctx context.Context, requestUUID string) (bool, error) {
	c.logger.Info("starting download and upload to GCS",
		zap.String("ApplicationName", constants.APPLICATION_NAME),
		zap.String("traceId", c.traceId))

	objectPath := filepath.Join(requestUUID, fileInfo.FileName)

	bucket := c.gcsClient.Bucket(c.bucketName)

	_, err := bucket.Object(objectPath).Attrs(ctx)
	if err != storage.ErrObjectNotExist {
		c.logger.Info("file already downloaded",
			zap.String("ApplicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.String("bucketName", c.bucketName),
			zap.String("fileUrl", fileInfo.FIleUrl),
			zap.String("fileName", fileInfo.FileName))

		return true, nil
	}
	if err != nil {
		c.logger.Info("error checking if file already exists",
			zap.String("ApplicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.String("bucketName", c.bucketName),
			zap.String("fileUrl", fileInfo.FIleUrl),
			zap.String("fileName", fileInfo.FileName),
			zap.Error(err))

		return false, fmt.Errorf("error checking object existence :%v", err)
	}

	c.logger.Info("file does not exists need to download and process file",
		zap.String("ApplicationName", constants.APPLICATION_NAME),
		zap.String("traceId", c.traceId),
		zap.String("bucketName", c.bucketName),
		zap.String("fileUrl", fileInfo.FIleUrl),
		zap.String("fileName", fileInfo.FileName))

	return false, nil
}

func (c *GCSClient) Close(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := c.gcsClient.Close()
	if err != nil {
		c.logger.Error("unable to close GCS client",
			zap.String("ApplicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.Error(err))
	}
	return nil
}
