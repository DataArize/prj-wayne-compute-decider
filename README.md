# Compute Decider

This service is a lightweight utility designed to **analyze files from remote URLs**. Given a URL, it performs a **HEAD request** to retrieve essential metadata such as **file extension**, **content type**, and **file size**—without downloading the entire file. This is particularly useful for services that need to make decisions based on file characteristics without incurring full data transfer costs.

---

## Features

- Accepts a `fileUrl` query parameter via HTTP GET
- Validates and parses remote file metadata using `HEAD` requests
- Extracts:
  - File extension (based on URL or Content-Type)
  - File size (from Content-Length header)
  - MIME type (from Content-Type header)
- Provides JSON response
- Clean logging using `zap`
- Minimal resource consumption — ideal for cloud-native workflows

---

## Tech Stack

- **Go 1.20+**
- **zap** for structured logging
- **Standard net/http libraries**
- Clean, modular code with support for context handling

---

## Getting Started

### Prerequisites

- Go 1.20 or higher
- Git

### Clone the Repository

```bash
git clone https://github.com/AmithSAI007/prj-wayne-compute-decider.git
cd prj-wayne-compute-decider
```

### Run Locally

To run the file analyzer, you can create a simple entry point or test using the following steps:

1. **Create a main.go (example usage)**

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AmithSAI007/prj-wayne-compute-decider.git/internal/processor"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	fileUrl := "https://example.com/sample.pdf"

	proc := processor.NewProcessor(fileUrl, logger)
	info := proc.AnalyzeFile(context.Background(), fileUrl, logger)

	if info.Error != "" {
		log.Fatalf("Error: %s\n", info.Error)
	}

	fmt.Printf("File Info:\nExtension: %s\nSize: %s\nContent Type: %s\n",
		info.FileExtension, info.FileSize, info.ContentType)
}
```

2. **Install Dependencies and Run**

```bash
go mod tidy
go run main.go
```

---

## Sample Output

```bash
File Info:
Extension: .pdf
Size: 34567
Content Type: application/pdf
```

---

## Logging

Logs are structured and emitted using `zap`:

```bash
INFO    Process completed     {"extension": ".pdf", "file size": "34567", "content type": "application/pdf"}
```
