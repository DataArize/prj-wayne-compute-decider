package gcs

import (
	"context"
	"errors"
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
		logger.With(zap.String("severity", "ERROR")).Error("unable to create storage client",
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
	objectPath := filepath.Join(requestUUID, fileInfo.FileName)
	c.logger.Info("checking if file already exists",
		zap.String("ApplicationName", constants.APPLICATION_NAME),
		zap.String("traceId", c.traceId),
		zap.String("bucketName", constants.HARDCODED_BUCKET_NAME),
		zap.String("objectPath", objectPath),
		zap.String("fileUrl", fileInfo.FIleUrl),
		zap.String("fileName", fileInfo.FileName))

	bucket := c.gcsClient.Bucket(constants.HARDCODED_BUCKET_NAME)

	_, err := bucket.Object(objectPath).Attrs(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		c.logger.Info("file does not exists start download",
			zap.String("ApplicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.String("objectPath", objectPath),
			zap.String("bucketName", constants.HARDCODED_BUCKET_NAME),
			zap.String("fileUrl", fileInfo.FIleUrl),
			zap.String("fileName", fileInfo.FileName))

		return false, nil
	}
	if err != nil {
		c.logger.Info("error checking if file already exists",
			zap.String("ApplicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.String("objectPath", objectPath),
			zap.String("bucketName", constants.HARDCODED_BUCKET_NAME),
			zap.String("fileUrl", fileInfo.FIleUrl),
			zap.String("fileName", fileInfo.FileName),
			zap.Error(err))

		return false, fmt.Errorf("error checking object existence :%v", err)
	}

	c.logger.Info("file already exists",
		zap.String("ApplicationName", constants.APPLICATION_NAME),
		zap.String("traceId", c.traceId),
		zap.String("objectPath", objectPath),
		zap.String("bucketName", constants.HARDCODED_BUCKET_NAME),
		zap.String("fileUrl", fileInfo.FIleUrl),
		zap.String("fileName", fileInfo.FileName))

	return true, nil
}

func (c *GCSClient) Close(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := c.gcsClient.Close()
	if err != nil {
		c.logger.With(zap.String("severity", "ERROR")).Error("unable to close GCS client",
			zap.String("ApplicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.Error(err))
	}
	return nil
}
