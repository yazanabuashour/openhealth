# OpenHealth Client

Use this skill when an agent needs local-first health data from an OpenHealth service through the generated Go client.

## When To Use It

- Read the OpenHealth API from Go code.
- Build agent-side integrations that need typed health summaries, history, trends, medications, or labs.
- Keep client code pinned to the OpenAPI-generated contract instead of hand-written HTTP calls.
- Use the local in-process runtime instead of binding a host port or running a background daemon.

## Install Surface

- Install the first tagged module release with `go get github.com/yazanabuashour/openhealth@v0.1.0`, then import `github.com/yazanabuashour/openhealth/client`.
- Import `github.com/yazanabuashour/openhealth/client`.
- Prefer `client.OpenLocal(client.LocalConfig{})` for the default user-machine install surface.
- Use `client.NewDefault(baseURL)` only when you intentionally want to talk to an explicit HTTP server.
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
  api, err := client.OpenLocal(client.LocalConfig{})
  if err != nil {
    log.Fatal(err)
  }
  defer api.Close()

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

- `client.OpenLocal(...)` opens SQLite, runs migrations, and serves the OpenAPI handler in-process.
- Default data location is `${XDG_DATA_HOME:-~/.local/share}/openhealth/openhealth.db`; override it with `client.LocalConfig`, `OPENHEALTH_DATA_DIR`, or `OPENHEALTH_DATABASE_PATH`.
- Run the server with `go run ./cmd/openhealth serve` only for maintainer/debug workflows.
- A fuller sample lives at `examples/client_summary/main.go`.
