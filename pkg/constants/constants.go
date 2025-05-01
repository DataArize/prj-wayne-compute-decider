package constants

const (
	APPLICATION_NAME     = "compute-decider"
	DATASET_ID           = "audit_layer"
	TABLE_ID             = "contact_trace_logs"
	CONTRACT_QUEUE_TABLE = "contract_file_queue"
	FILE_URL_PARAM       = "fileUrl"
	CONTENT_TYPE         = "Content-Type"
	RANGE_SUPPORTED      = "Accept-Ranges"
	APPLICATION_JSON     = "application/json"
	HEAD                 = "HEAD"
	CONTENT_LENGTH       = "Content-Length"
	FILE_SIZE_BYTES      = 1073741824.0
	REGION               = "us-central1"
	JOB_NAME             = "JOB_NAME"
	BYTES                = "bytes"

	// ENV CONSTANTS
	PROJECT_ID = "GCP_PROJECT_ID"

	// STATUS CONSTANTS
	STARTED     = "STARTED"
	COMPLETED   = "COMPLETED"
	IN_PROGRESS = "IN_PROGRESS"
	FAILED      = "FAILED"

	// EVENT CONSTANTS
	APPLICATION_STARTED_EVENT      = "compute_decider.application_started"
	REQUEST_BODY_FAILED            = "compute_decider.request_body_failed"
	INVALID_JSON_FORMAT            = "compute_decider.invalid_json_format"
	FILE_URL_MISSING               = "compute_decider.file_url_missing"
	ERROR_FETCHING_FILE_SIZE       = "compute_decider.error_fetching_file_size"
	TRIGGER_CLOUD_RUN_JOB          = "compute_decider.trigger_cloud_run_job"
	TRIGGER_CLOUD_BATCH_JOB        = "compute_decider.trigger_cloud_batch_job"
	FAILED_TRIGGER_CLOUD_RUN_JOB   = "compute_decider.trigger_cloud_run_job_failed"
	FAILED_TRIGGER_CLOUD_BATCH_JOB = "compute_decider.trigger_cloud_batch_job_failed"
	APPLICATION_COMPLETED_EVENT    = "compute_decider.application_completed"

	// MAX FILE SIZE
	MAX_FILE_SIZE = 25

	// FILE EXTENSIONS
	GZ  = ".gz"
	ZIP = ".zip"

	ENVIRONMENT = "DEV"

	// JOB NAME
	CLOUD_RUN_JOB_NAME      = "prj-wayne-file-streamer"
	ZIP_DOWNLOADER_JOB_NAME = "prj-wayne-zip-downloader"
	JOB_PREFIX              = "projects/%s/locations/%s/jobs/%s"
)
