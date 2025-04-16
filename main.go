package decider

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/processor"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	FILE_URL_PARAM   = "fileUrl"
	CONTENT_TYPE     = "Content-Type"
	APPLICATION_JSON = "application/json"
)

type RequestBody struct {
	FileUrl string `json:"fileUrl"`
}

func AnalyzeFileHandler(w http.ResponseWriter, r *http.Request) {
	traceId := uuid.New().String()
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	ctx := r.Context()

	logger.Info("Application started", zap.String("traceId", traceId))

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var requestData RequestBody
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	fileUrl := requestData.FileUrl

	if fileUrl == "" {
		http.Error(w, "Missing 'fileUrl' parameter", http.StatusBadRequest)
		logger.Error("Bad Request", zap.String("message", "Missing fileUrl Parameter"))
		return
	}

	processor := processor.NewProcessor(fileUrl, logger)

	result := processor.AnalyzeFile(ctx, fileUrl, logger)

	result.TraceId = traceId

	w.Header().Set(CONTENT_TYPE, APPLICATION_JSON)

	if result.Error != "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	json.NewEncoder(w).Encode(result)
}
