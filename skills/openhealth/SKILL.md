# OpenHealth Client

Use this skill when an agent needs local-first health data from an OpenHealth service through the generated Go client.

## When To Use It

- Read the OpenHealth API from Go code.
- Build agent-side integrations that need typed health summaries, history, trends, medications, or labs.
- Keep client code pinned to the OpenAPI-generated contract instead of hand-written HTTP calls.

## Install Surface

- Import `github.com/yazanabuashour/openhealth/client`.
- Prefer `client.NewDefault(baseURL)` for a ready-to-use client with timeout and safe retries for idempotent reads.
- Use `client.NewClientWithResponses(...)` directly only when custom transport wiring is required.

## Minimal Example

```go
package main

import (
  "context"
  "log"

  "github.com/yazanabuashour/openhealth/client"
)

func main() {
  api, err := client.NewDefault("http://localhost:8080")
  if err != nil {
    log.Fatal(err)
  }

  summary, err := api.GetHealthSummaryWithResponse(context.Background())
  if err != nil {
    log.Fatal(err)
  }

  if summary.JSON200 == nil {
    log.Fatalf("unexpected status: %s", summary.Status())
  }

  log.Printf("active medications=%d", summary.JSON200.ActiveMedicationCount)
}
```

## Notes

- Run the server with `go run ./cmd/openhealth serve`.
- Run migrations first with `go run ./cmd/openhealth migrate`.
- A fuller sample lives at `examples/client_summary/main.go`.
