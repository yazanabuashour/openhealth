# Weight Task Recipes

Use these snippets after opening the local runtime:

```go
api, err := client.OpenLocal(client.LocalConfig{})
if err != nil {
	log.Fatal(err)
}
defer api.Close()
ctx := context.Background()
```

`client.OpenLocal(client.LocalConfig{})` honors the configured local
environment, including `OPENHEALTH_DATABASE_PATH`. Do not search for the
database path before using these snippets; run the SDK helper first.

## Add Or Reapply Weights

Use `UpsertWeight` for natural-language data-entry requests. It returns
`created`, `already_exists`, or `updated`.

```go
recordedAt, err := time.Parse(time.DateOnly, "2026-03-29")
if err != nil {
	log.Fatal(err)
}

result, err := api.UpsertWeight(ctx, client.WeightRecordInput{
	RecordedAt: recordedAt,
	Value:      152.2,
	Unit:       client.WeightUnitLb,
})
if err != nil {
	log.Fatal(err)
}

log.Printf("%s %.1f lb %s", result.Entry.RecordedAt.Format(time.DateOnly), result.Entry.Value, result.Status)
```

## Add Only If New

Use `RecordWeight` when duplicates should fail instead of being treated as
idempotent.

```go
entry, err := api.RecordWeight(ctx, client.WeightRecordInput{
	RecordedAt: recordedAt,
	Value:      152.2,
	Unit:       client.WeightUnitLb,
})
if err != nil {
	log.Fatal(err)
}
log.Printf("created %s %.1f lb", entry.RecordedAt.Format(time.DateOnly), entry.Value)
```

## Latest And History

```go
latest, err := api.LatestWeight(ctx)
if err != nil {
	log.Fatal(err)
}
if latest == nil {
	log.Printf("no weight history in %s", api.Paths.DatabasePath)
	return
}

weights, err := api.ListWeights(ctx, client.WeightListOptions{Limit: 25})
if err != nil {
	log.Fatal(err)
}
for _, weight := range weights {
	log.Printf("%s %.1f %s", weight.RecordedAt.Format(time.DateOnly), weight.Value, weight.Unit)
}
```

## Bounded History

Use `WeightListOptions{From: ..., To: ...}` when the user asks for a specific
date range. Parse the start date, parse the end date, and make `To` inclusive
through the end of that day.

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

When reporting the result, mirror every row in the bounded query output, newest
first. Do not report only the latest row, and do not include older or newer
records that were not part of the requested range. Do not mention excluded dates
at all, even to say they were excluded.

For example, if the requested range is `2026-03-29` through `2026-03-30` and the
bounded query returns `2026-03-30` and `2026-03-29`, report both rows.

If a Go run is unavailable and you inspect SQLite directly as a fallback, keep
the same date-only bounds:

```sql
SELECT substr(recorded_at, 1, 10) AS date, value, unit
FROM health_weight_entry
WHERE deleted_at IS NULL
  AND substr(recorded_at, 1, 10) >= '2026-03-29'
  AND substr(recorded_at, 1, 10) <= '2026-03-30'
ORDER BY recorded_at DESC, id DESC;
```

## Date Handling

If a user gives a short date like `03/29`, resolve the year from the conversation
or ask for it when the context is ambiguous. Pass OpenHealth an explicit
`YYYY-MM-DD` date converted with `time.Parse(time.DateOnly, value)`.
