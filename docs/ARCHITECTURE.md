# TECHNICAL SPECIFICATION

## Table of Contents

- [Overview](#overview)
- [System Goals](#system-goals)
- [High-Level Architecture](#high-level-architecture)
- [Component Responsibilities](#component-responsibilities)
  - [1. Compute-Decider](#1-compute-decider)
  - [2. File-Streamer](#2-file-streamer)
  - [3. Zip-Downloader](#3-zip-downloader)
- [Workflow Summary](#workflow-summary)
- [Error Handling & Observability](#error-handling--observability)
- [Audit Logging](#audit-logging)
- [Environment Variables](#environment-variables)
- [Security Considerations](#security-considerations)

---

## Overview

The Wayne Compute Pipeline is a **serverless, event-driven system** for processing large data files. It classifies incoming files and decides whether to stream them into storage directly or submit them for batch decompression. The pipeline leverages **Cloud Functions**, **Cloud Run**, **BigQuery**, and **GCS** to offer a highly observable and scalable solution.

---

## System Goals

- Automate classification and routing of files for processing
- Reduce manual operations and error-prone file handling
- Maintain full traceability and audit logging
- Optimize compute resource usage based on file type and size
- Integrate cleanly into broader data pipelines

---

## High-Level Architecture

![Architecture Diagram](https://github.com/user-attachments/assets/1edf4b75-e3e7-4640-8a22-74b95912b9e4)

---

## Flow diargam

## ![Flow Diagram](https://github.com/user-attachments/assets/a188d253-8142-49f8-b142-baba9285862f)

## Component Responsibilities

### 1. Compute-Decider

- **Type**: Google Cloud Function
- **Purpose**: Entry point for file analysis and routing
- **Responsibilities**:
  - Validate and parse HTTP requests containing fileUrl[]
  - Issue HEAD requests to check file metadata (size, extension, etc.)
  - Log events to BigQuery
  - Trigger Cloud Run jobs based on rules: - .gz → File-Streamer - .zip → insert job into BQ Queue, then trigger Zip-Downloader
- **Audit Events**:
  - `APPLICATION_STARTED_EVENT`
  - `FILE_URL_MISSING`
  - `INVALID_JSON_FORMAT`
  - `TRIGGER_CLOUD_RUN_JOB`
  - `TRIGGER_CLOUD_BATCH_JOB`

---

### 2. File-Streamer

- **Type**: Cloud Run Job (Go)
- **Purpose**: Downloads and streams .gz files into GCS after decompressing in-memory
- **Responsibilities**:
  - Accepts traceId and fileUrl as arguments
  - Fetches file via HTTP
  - Decompresses .gz file using pgzip
  - Uploads result to GCS bucket
  - Audit Events:
    - `APPLICATION_STARTED_EVENT`
    - `GCS_CLIENT_CREATION_FAILED`
    - `UNABLE_TO_CREATE_GZIP_READER`
    - `WRITE_FAILED`
    - `APPLICATION_COMPLETED_EVENT`

---

### 3. Zip-Downloader

- **Type**: Cloud Run Job (Go)
- **Purpose**: Downloads .zip files and stores them in a mount path (e.g., GCS FUSE)
- **Responsibilities**:
  - Accepts traceId, fileUrl, and fileName as arguments
  - Performs HTTP GET to download the .zip file
  - Saves it to mounted path (e.g., /mnt/data/<traceId>/<fileName>)
  - Audit Events:
    - `APPLICATION_STARTED_EVENT`
    - `GCS_CLIENT_CREATION_FAILED`
    - `UNABLE_TO_CREATE_DEST_PATH`
    - `INVALID_RESPONSE_FROM_THE_API`
    - `APPLICATION_COMPLETED_EVENT`

---

## Workflow Summary

- For `.gz` Files
- Client sends HTTP request to Compute-Decider.
- Compute-Decider detects .gz extension.
- Triggers Cloud Run Job: File-Streamer.
- File-Streamer:

  - Downloads the file
  - Decompresses in-memory
  - Streams contents to GCS
  - BigQuery logs audit trail.

- For `.zip` Files
- Client sends HTTP request to Compute-Decider.
- Compute-Decider detects .zip extension.
- Inserts a row into the Contract File Queue (BQ Table).
- Triggers Cloud Run Job: Zip-Downloader.
- Zip-Downloader:
  - Downloads the file
  - Saves it to mounted path
  - BigQuery logs audit trail.

---

## Error Handling & Observability

- All services use Zap Logger with structured logging:
  - Includes `traceId`, `fileUrl`, `ApplicationName`, and `error` messages.
- Failures are logged with:
  - Clear event names
  - Error context and trace IDs
- Each error is also pushed to BigQuery Audit Table via:
  - `LogAuditData()` in all services

---

## Audit Logging

**BigQuery Table Schema** (simplified):

| Field        | Type      | Description                                    |
| ------------ | --------- | ---------------------------------------------- |
| TraceID      | STRING    | Correlates logs across services                |
| ContractID   | STRING    | Same as TraceID                                |
| Event        | STRING    | Event name (e.g., `APPLICATION_STARTED_EVENT`) |
| Status       | STRING    | STARTED, FAILED, COMPLETED                     |
| Timestamp    | TIMESTAMP | Event time                                     |
| FunctionName | STRING    | The service name                               |
| Message      | STRING    | Additional context                             |

## Environment Variables

| Name          | Required | Used In         | Description              |
| ------------- | -------- | --------------- | ------------------------ |
| `PROJECT_ID`  | True     | All apps        | GCP Project ID           |
| `BUCKET_NAME` | True     | File & Zip      | Target GCS bucket name   |
| `JOB_NAME`    | True     | Compute-Decider | Cloud Run job to trigger |
| `REGION`      | True     | Compute-Decider | GCP Region               |

---

## Security Considerations

- All HTTP entry points (Cloud Function and FetchAPI) must be protected via:
  - VPC-SC or IAM-based access control
  - Optional: Google Cloud API Gateway
- File streaming uses HEAD/GET only (no POSTs or downloads to local disks)
- Trace IDs are UUIDs to prevent correlation leakage
- BigQuery tables can be access-controlled via standard IAM policies

---

## Summary

This technical specification documents the Wayne Compute Pipeline — a modular, serverless, event-driven file processing system. Each component is independently deployable, observable, and auditable. The system ensures efficient processing and classification of large files using Google Cloud-native services.
