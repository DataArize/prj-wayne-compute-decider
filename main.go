// Package decider contains the HTTP handler for analyzing file URLs.
// It serves as an entry point for the Cloud Function and coordinates logging,
// validation, and delegation to internal components.
package decider

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/bigquery"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/compute"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/gcs"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/model"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/processor"
	"github.com/AmithSAI007/prj-wayne-compute-decider.git/pkg/constants"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AnalyzeFileHandler is the main HTTP handler function for the Cloud Function.
// It validates the incoming request, initializes required clients, logs audit events,
// and delegates file analysis to the processor. Results are returned as a JSON response.
func AnalyzeFileHandler(w http.ResponseWriter, r *http.Request) {
	traceId := uuid.New().String()

	// Initialize production logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	ctx := r.Context()

	logger.Info("Application started",
		zap.String("applicationName", constants.APPLICATION_NAME),
		zap.String("traceId", traceId))

	// Load configuration from environment variables and constants
	projectId := os.Getenv(constants.PROJECT_ID)
	bucketName := os.Getenv(constants.BUCKET_NAME)
	projectRegion := constants.REGION
	jobName := os.Getenv(constants.JOB_NAME)
	if projectId == "" {
		logger.Error("project Id not specified",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))

		http.Error(w, "project id not specified", http.StatusBadRequest)
		return

	}

	if r.Method == http.MethodGet && r.URL.Path == constants.HEALTH {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	// Initialize BigQuery client
	client, err := bigquery.NewClient(ctx, logger, projectId, traceId)
	if err != nil {
		logger.Error("biquery client creation failed",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))

		http.Error(w, "bigquery client creation failed", http.StatusBadRequest)
		return
	}

	// Initialize Compute client
	compute, err := compute.NewCompute(ctx, logger, traceId)
	if err != nil {
		logger.Error("unable to create cloud run job client",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))

		http.Error(w, "unable to create cloud run job client", http.StatusBadRequest)
		return
	}

	// Log application start event
	client.LogAuditData(ctx, model.AuditEvent{
		TraceID:      traceId,
		ContractId:   traceId,
		Event:        constants.APPLICATION_STARTED_EVENT,
		Status:       constants.STARTED,
		Timestamp:    time.Now(),
		FunctionName: constants.APPLICATION_NAME,
		Message:      "application started",
	})

	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed to read request body",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))
		client.LogAuditData(ctx, model.AuditEvent{
			TraceID:      traceId,
			ContractId:   traceId,
			Event:        constants.REQUEST_BODY_FAILED,
			Status:       constants.FAILED,
			Timestamp:    time.Now(),
			FunctionName: constants.APPLICATION_NAME,
			Message:      err.Error(),
		})

		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Unmarshal request JSON into a structured format
	var requestData model.RequestBody
	if err := json.Unmarshal(body, &requestData); err != nil {
		logger.Error("invalid JSON format",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))

		client.LogAuditData(ctx, model.AuditEvent{
			TraceID:      traceId,
			ContractId:   traceId,
			Event:        constants.INVALID_JSON_FORMAT,
			Status:       constants.FAILED,
			Timestamp:    time.Now(),
			FunctionName: constants.APPLICATION_NAME,
			Message:      err.Error(),
		})

		return
	}

	fileUrl := requestData.FileUrl
	requestUUID := requestData.RequestUUID

	if len(fileUrl) == 0 {
		logger.Error("Bad Request",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.String("message", "Missing fileUrl Parameter"))

		client.LogAuditData(ctx, model.AuditEvent{
			TraceID:      traceId,
			ContractId:   traceId,
			Event:        constants.FILE_URL_MISSING,
			Status:       constants.FAILED,
			Timestamp:    time.Now(),
			FunctionName: constants.APPLICATION_NAME,
			Message:      "missing fileUrl Parameter",
		})

		http.Error(w, "Missing 'fileUrl' parameter", http.StatusBadRequest)
		return
	}

	gcsClient, err := gcs.NewGCSClient(logger, bucketName, traceId, ctx)
	if err != nil {
		logger.Error("error creating GCS client",
			zap.String("applicationName", constants.APPLICATION_NAME),
			zap.String("traceId", traceId),
			zap.Error(err))

		client.LogAuditData(ctx, model.AuditEvent{
			TraceID:      traceId,
			ContractId:   traceId,
			Event:        constants.ERROR_CREATING_GCS_CLIENT,
			Status:       constants.FAILED,
			Timestamp:    time.Now(),
			FunctionName: constants.APPLICATION_NAME,
			Message:      err.Error(),
		})

		http.Error(w, "error fetching file size", http.StatusInternalServerError)
		return

	}

	// Instantiate processor and analyze the file
	processor := processor.NewProcessor(traceId, fileUrl, logger, client, compute, projectId, projectRegion, jobName, gcsClient)

	result := processor.AnalyzeFileUrls(ctx, fileUrl, requestUUID)

	w.Header().Set(constants.CONTENT_TYPE, constants.APPLICATION_JSON)

	// Handle any errors from processing
	for _, res := range result {
		if res.Error != "" {
			logger.Error("error fetching file size",
				zap.String("applicationName", constants.APPLICATION_NAME),
				zap.String("traceId", traceId),
				zap.String("error", res.Error))

			client.LogAuditData(ctx, model.AuditEvent{
				TraceID:      traceId,
				ContractId:   traceId,
				FileUrl:      res.FIleUrl,
				Event:        constants.ERROR_FETCHING_FILE_SIZE,
				Status:       constants.FAILED,
				Timestamp:    time.Now(),
				FunctionName: constants.APPLICATION_NAME,
				Message:      res.Error,
			})

			http.Error(w, "error fetching file size", http.StatusInternalServerError)
			return
		}
	}

	// Respond with result
	json.NewEncoder(w).Encode(result)

	// Log application completion event
	client.LogAuditData(ctx, model.AuditEvent{
		TraceID:      traceId,
		ContractId:   traceId,
		Event:        constants.APPLICATION_COMPLETED_EVENT,
		Status:       constants.COMPLETED,
		Timestamp:    time.Now(),
		FunctionName: constants.APPLICATION_NAME,
		Message:      "application completed",
	})

	logger.Info("process completed",
		zap.String("traceId", traceId),
		zap.String("applicationName", constants.APPLICATION_NAME))
}
