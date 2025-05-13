// Package compute provides a wrapper around Google Cloud Run jobs.
// It is responsible for triggering cloud run jobs with appropriate arguments
// and managing the lifecycle of the Run client.
package compute

import (
	"context"
	"fmt"
	"time"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/pkg/constants"
	"go.uber.org/zap"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
)

// Compute wraps the Cloud Run JobsClient with logging and traceability context.
type Compute struct {
	logger  *zap.Logger     // Logger for structured logging
	client  *run.JobsClient // Google Cloud Run jobs client
	traceId string          // Trace ID for request tracking
}

// NewCompute initializes and returns a new Compute instance.
// It creates a Cloud Run JobsClient and sets up logging with trace information.
func NewCompute(ctx context.Context, logger *zap.Logger, traceId string) (*Compute, error) {
	client, err := run.NewJobsClient(ctx)
	if err != nil {
		logger.With(zap.String("severity", "ERROR")).Error("unable to create cloud run job client",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))
		return nil, err
	}
	return &Compute{
		logger:  logger,
		client:  client,
		traceId: traceId,
	}, err
}

// TriggerFileStreamerJob starts a Cloud Run job using the provided project,
// region, job name, and arguments. It logs both the initiation and result
// of the operation for observability and debugging.
func (c *Compute) TriggerFileStreamerJob(ctx context.Context, projectId string, region string, jobName string, args []string) error {
	name := fmt.Sprintf(constants.JOB_PREFIX, projectId, region, jobName)
	c.logger.Info("attempting to trigger cloud run job",
		zap.String("applicationName", constants.APPLICATION_NAME),
		zap.String("traceId", c.traceId),
		zap.String("region", region),
		zap.String("jobName", constants.CLOUD_RUN_JOB_NAME),
		zap.String("name", name))

	req := &runpb.RunJobRequest{
		Name: name,
		Overrides: &runpb.RunJobRequest_Overrides{
			ContainerOverrides: []*runpb.RunJobRequest_Overrides_ContainerOverride{
				{
					Args: args,
				},
			},
			TaskCount: 1,
		},
	}

	op, err := c.client.RunJob(ctx, req)
	if err != nil {
		c.logger.With(zap.String("severity", "ERROR")).Error("failed to trigger cloud run job",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.Error(err))

		return err
	}

	c.logger.Info("triggered cloud run job",
		zap.String("applicationName", constants.APPLICATION_NAME),
		zap.String("traceId", c.traceId),
		zap.String("jobName", op.Name()))

	return nil
}

// Close gracefully closes the Cloud Run JobsClient to free resources.
func (c *Compute) Close(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := c.client.Close()
	if err != nil {
		c.logger.With(zap.String("severity", "ERROR")).Error("unable to close cloud run job client",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.Error(err))
		return err
	}
	return nil
}
