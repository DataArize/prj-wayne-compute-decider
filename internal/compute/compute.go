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

type Compute struct {
	logger  *zap.Logger
	client  *run.JobsClient
	traceId string
}

func NewCompute(ctx context.Context, logger *zap.Logger, traceId string) (*Compute, error) {
	client, err := run.NewJobsClient(ctx)
	if err != nil {
		logger.Error("unable to create cloud run job client",
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

func (c *Compute) TriggerFileStreamerJob(ctx context.Context, projectId string, region string, jobName string, args []string) error {
	c.logger.Info("attempting to trigger cloud run job",
		zap.String("applicationName", constants.APPLICATION_NAME),
		zap.String("traceId", c.traceId),
		zap.String("jobName", constants.CLOUD_RUN_JOB_NAME))

	req := &runpb.RunJobRequest{
		Name: fmt.Sprintf(constants.JOB_PREFIX, projectId, region, jobName),
		Overrides: &runpb.RunJobRequest_Overrides{
			ContainerOverrides: []*runpb.RunJobRequest_Overrides_ContainerOverride{
				{
					Name: constants.CLOUD_RUN_JOB_NAME,
					Args: args,
				},
			},
			TaskCount: 1,
		},
	}

	op, err := c.client.RunJob(ctx, req)
	if err != nil {
		c.logger.Error("failed to trigger cloud run job",
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

func (c *Compute) Close(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := c.client.Close()
	if err != nil {
		c.logger.Error("unable to close cloud run job client",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", c.traceId),
			zap.Error(err))
		return err
	}
	return nil
}
