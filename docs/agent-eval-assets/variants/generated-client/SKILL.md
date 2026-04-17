---
name: generated-client
description: Use this skill when an agent needs the generated OpenHealth Go client and local-first runtime to read or write local health data through github.com/yazanabuashour/openhealth/client.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and use of github.com/yazanabuashour/openhealth/client.
---

# OpenHealth Generated Client Baseline

Use this isolated baseline when evaluating the older generated-client skill
surface. It intentionally does not mention the hand-written weight helper
methods, so it can be compared against the production SDK-oriented skill.

## When To Use It

- Read or write OpenHealth API data from Go code.
- Build agent-side integrations that need typed health summaries, history,
  trends, medications, labs, or weight data.
- Keep client code pinned to the OpenAPI-generated contract instead of
  hand-written HTTP calls.
- Use the local in-process runtime instead of binding a host port or running a
  background daemon.

## Install Surface

- Install from the current development line until a release tag exists:
  `go get github.com/yazanabuashour/openhealth@main`.
- Import `github.com/yazanabuashour/openhealth/client`.
- Prefer `client.OpenLocal(client.LocalConfig{})` for the default user-machine
  install surface.
- Use `client.NewDefault(baseURL)` only when you intentionally want to talk to
  an explicit HTTP server.
- Use `client.NewClientWithResponses(...)` directly only when custom transport
  wiring is required.

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

## Weight History

To answer "what is my weight history?", call the generated weight list
endpoint. Results are ordered newest first, so the latest weight is `items[0]`
when the list is non-empty.

```go
limit := client.Limit(25)
weights, err := api.ListHealthWeightWithResponse(ctx, &client.ListHealthWeightParams{
	Limit: &limit,
})
if err != nil {
	log.Fatal(err)
}
if weights.JSON200 == nil {
	log.Fatalf("unexpected status: %s", weights.Status())
}
if len(weights.JSON200.Items) == 0 {
	log.Printf("no weight history in %s", api.Paths.DatabasePath)
	return
}

latest := weights.JSON200.Items[0]
log.Printf("latest weight: %.1f %s at %s", latest.Value, latest.Unit, latest.RecordedAt)
```

Use `From`, `To`, and `Limit` on `client.ListHealthWeightParams` for date
filtering. Use `GetHealthWeightTrendWithResponse` only when the user asks for
trend or chart-style data; use `ListHealthWeightWithResponse` for history and
latest-weight questions.

For a bounded date range, parse both dates, make `To` inclusive through the end
of the requested final day, and report every row returned by that bounded
request, newest first. Do not report only the latest row, and do not append
older or newer rows from the database to the final answer. Do not mention
excluded dates at all, even to say they were excluded.

```go
fromDate, err := time.Parse(time.DateOnly, "2026-03-29")
if err != nil {
	log.Fatal(err)
}
toDate, err := time.Parse(time.DateOnly, "2026-03-30")
if err != nil {
	log.Fatal(err)
}
toEnd := toDate.Add(24*time.Hour - time.Nanosecond)

weights, err := api.ListHealthWeightWithResponse(ctx, &client.ListHealthWeightParams{
	From: &fromDate,
	To:   &toEnd,
})
if err != nil {
	log.Fatal(err)
}
if weights.JSON200 == nil {
	log.Fatalf("unexpected status: %s", weights.Status())
}
for _, weight := range weights.JSON200.Items {
	log.Printf("%s %.1f %s", weight.RecordedAt.Format(time.DateOnly), weight.Value, weight.Unit)
}
```

For the range `2026-03-29` through `2026-03-30`, a result containing
`2026-03-30` and `2026-03-29` must be reported as both rows, newest first.
If a Go run is unavailable and you inspect SQLite directly as a fallback, keep
the same date-only bounds instead of listing all history:

```sql
SELECT substr(recorded_at, 1, 10) AS date, value, unit
FROM health_weight_entry
WHERE deleted_at IS NULL
  AND substr(recorded_at, 1, 10) >= '2026-03-29'
  AND substr(recorded_at, 1, 10) <= '2026-03-30'
ORDER BY recorded_at DESC, id DESC;
```

## Write Weight Entries

For a generated-client-only write, call the generated create endpoint with an
explicit timestamp, value, and unit.

Before writing, reject or clarify unsafe inputs:

- If the user gives a short date without explicit year context, ask which year
  they mean and do not write.
- If the value is zero or negative, do not write.
- If the unit is not pounds/`lb`, do not convert it and do not write.

```go
recordedAt, err := time.Parse(time.DateOnly, "2026-03-29")
if err != nil {
	log.Fatal(err)
}

created, err := api.CreateHealthWeightWithResponse(ctx, client.CreateHealthWeightJSONRequestBody{
	RecordedAt: recordedAt,
	Value:      152.2,
	Unit:       client.CreateHealthWeightRequestUnitLb,
})
if err != nil {
	log.Fatal(err)
}
if created.JSON201 == nil {
	log.Fatalf("unexpected status: %s", created.Status())
}
```

## Notes

- `client.OpenLocal(...)` opens SQLite, runs migrations, and serves the OpenAPI
  handler in-process.
- Default data location is `${XDG_DATA_HOME:-~/.local/share}/openhealth/openhealth.db`;
  override it with `client.LocalConfig`, `OPENHEALTH_DATA_DIR`, or
  `OPENHEALTH_DATABASE_PATH`.
- Run `go run ./cmd/openhealth serve` only for maintainer or debug workflows.
