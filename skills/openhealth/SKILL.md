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
- Open local data with `client.OpenLocal(client.LocalConfig{})`. It honors the
  configured local environment, including `OPENHEALTH_DATABASE_PATH`, so do not
  search for the database path unless `OpenLocal` fails.
- For routine weight tasks, start with this skill and
  [references/weights.md](references/weights.md). Use `UpsertWeight`,
  `RecordWeight`, `ListWeights`, and `LatestWeight`; inspect `client/weight.go`
  only when the snippets do not answer the task.
- Use `client.LocalConfig{DatabasePath: "..."}` or
  `client.LocalConfig{DataDir: "..."}` only when the user names a specific
  database or you are using a temp test database.

Do not inspect `client.gen.go`, generated server code, the Go module cache, or
large dependency directories for routine add/list/latest weight tasks. Use
targeted repo searches only when the SDK facade does not cover the user's ask.
For routine weight tasks, avoid repo-wide file listings or content searches like
`rg --files .` or `rg -n ... -S .`; search this skill, `references/weights.md`,
and `client/weight.go` directly when additional context is needed. Do not append
`.` to an otherwise targeted search, and do not search for `go.mod` or `go.sum`;
the eval starts in the repository root.

## Routine Weight Fast Path

For weight add, reapply, correction, latest, history, or bounded-range requests:

1. Read [references/weights.md](references/weights.md).
2. Copy the matching SDK helper snippet into a short temporary Go program.
3. Run it from the repository with the inherited environment.
4. Answer only from the program output and the requested user range.

Do not enumerate the repository, inspect generated files, inspect the Go module
cache, or query SQLite directly unless the SDK helper run fails.

For invalid input or ambiguous short-date requests, answer from the validation
rules in this skill. Do not search the repository or run code before rejecting
non-positive values, unsupported units, or dates that lack a year.

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

For a bounded history request, parse both dates and pass an inclusive end of
day for `To` so entries on the requested end date are included. Answer with
every row returned by the bounded query, newest first. Do not report only the
latest row, and do not append older or newer rows you may have seen while
inspecting the database. Do not mention excluded dates at all, even to say they
were excluded.

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

weights, err := api.ListWeights(ctx, client.WeightListOptions{
	From: &fromDate,
	To:   &toEnd,
})
if err != nil {
	log.Fatal(err)
}
for _, weight := range weights {
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

Weight entries are returned newest first. A focused reference with copyable
task snippets lives at [references/weights.md](references/weights.md).

## Generated Client Fallback

The generated OpenAPI client remains available for advanced API-contract work,
HTTP-server calls, or endpoints not yet covered by the SDK facade. Do not start
there for common agent tasks.
