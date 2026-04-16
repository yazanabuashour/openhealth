---
name: openhealth
description: Use this skill when an agent needs to read or write local-first OpenHealth data through the ergonomic Go SDK in github.com/yazanabuashour/openhealth/client.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and use of github.com/yazanabuashour/openhealth/client.
---

# OpenHealth Agent SDK

Use this skill for local-first OpenHealth tasks. The production agent path is
the hand-written Go SDK facade on top of `client.OpenLocal(...)`, not the raw
generated OpenAPI methods.

## Default Path

- Install from the current development line until a release tag exists:
  `go get github.com/yazanabuashour/openhealth@main`.
- Import `github.com/yazanabuashour/openhealth/client`.
- Open local data with `client.OpenLocal(client.LocalConfig{})`.
- Use `UpsertWeight`, `RecordWeight`, `ListWeights`, and `LatestWeight` for
  weight tasks.
- Use `client.LocalConfig{DatabasePath: "..."}` or
  `client.LocalConfig{DataDir: "..."}` only when the user names a specific
  database or you are using a temp test database.

Do not inspect `client.gen.go`, generated server code, the Go module cache, or
large dependency directories for routine add/list/latest weight tasks. Use
targeted repo searches only when the SDK facade does not cover the user's ask.

## Add Weight Entries

Prefer `UpsertWeight` for user-entered measurements. It is idempotent for the
same recorded date and unit, returns a status, and avoids duplicate manual
entries.

Before writing a measurement, validate the user's input exactly:

- Do not infer a year for short dates unless the user or conversation provides
  explicit year context. Ask for the year instead of writing.
- Do not convert unsupported units. V1 accepts pounds only, represented as
  `client.WeightUnitLb`.
- Do not write non-positive, missing, or otherwise invalid values. Tell the user
  the entry is invalid instead.

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/yazanabuashour/openhealth/client"
)

func main() {
	api, err := client.OpenLocal(client.LocalConfig{})
	if err != nil {
		log.Fatal(err)
	}
	defer api.Close()

	ctx := context.Background()
	for _, item := range []struct {
		date  string
		value float64
	}{
		{"2026-03-29", 152.2},
		{"2026-03-30", 151.6},
	} {
		recordedAt, err := time.Parse(time.DateOnly, item.date)
		if err != nil {
			log.Fatal(err)
		}
		result, err := api.UpsertWeight(ctx, client.WeightRecordInput{
			RecordedAt: recordedAt,
			Value:      item.value,
			Unit:       client.WeightUnitLb,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s %.1f lb %s", result.Entry.RecordedAt.Format(time.DateOnly), result.Entry.Value, result.Status)
	}
}
```

Use `RecordWeight` only when duplicate entries should fail with a conflict.

## Read Weight Entries

```go
latest, err := api.LatestWeight(ctx)
if err != nil {
	log.Fatal(err)
}
if latest == nil {
	log.Printf("no weight history in %s", api.Paths.DatabasePath)
	return
}
log.Printf("latest weight: %.1f %s at %s", latest.Value, latest.Unit, latest.RecordedAt.Format(time.RFC3339))

weights, err := api.ListWeights(ctx, client.WeightListOptions{Limit: 25})
if err != nil {
	log.Fatal(err)
}
for _, weight := range weights {
	log.Printf("%s %.1f %s", weight.RecordedAt.Format(time.DateOnly), weight.Value, weight.Unit)
}
```

Weight entries are returned newest first. A focused reference with copyable
task snippets lives at [references/weights.md](references/weights.md).

## Generated Client Fallback

The generated OpenAPI client remains available for advanced API-contract work,
HTTP-server calls, or endpoints not yet covered by the SDK facade. Do not start
there for common agent tasks.
