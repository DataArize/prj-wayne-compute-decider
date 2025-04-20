package decider

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/bigquery"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/model"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/processor"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/pkg/constants"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func AnalyzeFileHandler(w http.ResponseWriter, r *http.Request) {
	traceId := uuid.New().String()
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	ctx := r.Context()

	logger.Info("Application started",
		zap.String("applicationName", constants.APPLICATION_NAME),
		zap.String("traceId", traceId))

	projectId := os.Getenv(constants.PROJECT_ID)

	client, err := bigquery.NewClient(ctx, logger, projectId, traceId)
	if err != nil {
		http.Error(w, "bigquery client creation failed", http.StatusBadRequest)
		logger.Error("biquery client creation failed",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))
		return
	}

	client.LogAuditData(ctx, model.AuditEvent{
		TraceID:   traceId,
		Event:     constants.APPLICATION_STARTED_EVENT,
		Status:    constants.STARTED,
		Timestamp: time.Now(),
	})

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		logger.Error("failed to read request body",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))
		client.LogAuditData(ctx, model.AuditEvent{
			TraceID:   traceId,
			Event:     constants.REQUEST_BODY_FAILED,
			Status:    constants.FAILED,
			Timestamp: time.Now(),
		})

		return
	}
	defer r.Body.Close()

	var requestData model.RequestBody
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		logger.Error("invalid JSON format",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))

		client.LogAuditData(ctx, model.AuditEvent{
			TraceID:   traceId,
			Event:     constants.INVALID_JSON_FORMAT,
			Status:    constants.FAILED,
			Timestamp: time.Now(),
		})

		return
	}

	fileUrl := requestData.FileUrl

	if len(fileUrl) == 0 {
		http.Error(w, "Missing 'fileUrl' parameter", http.StatusBadRequest)
		logger.Error("Bad Request",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.String("message", "Missing fileUrl Parameter"))

		client.LogAuditData(ctx, model.AuditEvent{
			TraceID:   traceId,
			Event:     constants.FILE_URL_MISSING,
			Status:    constants.FAILED,
			Timestamp: time.Now(),
		})

		return
	}

	processor := processor.NewProcessor(traceId, fileUrl, logger)

	result := processor.AnalyzeFileUrls(ctx, fileUrl)

	w.Header().Set(constants.CONTENT_TYPE, constants.APPLICATION_JSON)

	for _, res := range result {
		if res.Error != "" {
			logger.Error("error fetching file size",
				zap.String("error", res.Error))

			client.LogAuditData(ctx, model.AuditEvent{
				TraceID:   traceId,
				Event:     constants.ERROR_FETCHING_FILE_SIZE,
				Status:    constants.FAILED,
				Timestamp: time.Now(),
			})

		}
	}

	json.NewEncoder(w).Encode(result)

	client.LogAuditData(ctx, model.AuditEvent{
		TraceID:   traceId,
		Event:     constants.APPLICATION_COMPLETED_EVENT,
		Status:    constants.COMPLETED,
		Timestamp: time.Now(),
	})

	logger.Info("process completed",
		zap.String("traceId", traceId),
		zap.String("applicationName", constants.APPLICATION_NAME))
}
