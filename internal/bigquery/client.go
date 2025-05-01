package bigquery

import (
	"context"
	"fmt"
	"time"

	bq "cloud.google.com/go/bigquery"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/model"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/pkg/constants"

	"go.uber.org/zap"
)

type Client struct {
	logger    *zap.Logger
	traceId   string
	projectId string
	client    *bq.Client
}

func NewClient(ctx context.Context, logger *zap.Logger, projectId string, traceId string) (*Client, error) {
	client, err := bq.NewClient(ctx, projectId)
	if err != nil {
		logger.Error("unable to create bigquery client",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))

		return nil, fmt.Errorf("unable to create bigquery client: %v", err)
	}

	return &Client{
		logger:    logger,
		traceId:   traceId,
		projectId: projectId,
		client:    client,
	}, nil
}

func (c *Client) LogAuditData(ctx context.Context, event model.AuditEvent) error {
	inserter := c.client.Dataset(constants.DATASET_ID).Table(constants.TABLE_ID).Inserter()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	err := inserter.Put(ctx, []*model.AuditEvent{&event})
	if err != nil {
		c.logger.Info("unable to persist data into bigquery",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.Error(err))
	}
	return nil
}

func (c *Client) ContractFileQueue(ctx context.Context, event model.ContractFileEvent) error {
	inserter := c.client.Dataset(constants.DATASET_ID).Table(constants.CONTRACT_QUEUE_TABLE).Inserter()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	err := inserter.Put(ctx, []*model.ContractFileEvent{&event})
	if err != nil {
		c.logger.Info("unable to persist data into bigquery",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.Error(err))
	}
	return nil
}

func (c *Client) Close(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := c.client.Close()
	if err != nil {
		c.logger.Info("unable to close bigquery client",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.Error(err))
		return fmt.Errorf("unable to close bigquery client: %v", err)
	}
	return nil
}
