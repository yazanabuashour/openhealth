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

## Date Handling

If a user gives a short date like `03/29`, resolve the year from the conversation
or ask for it when the context is ambiguous. Pass OpenHealth an explicit
`YYYY-MM-DD` date converted with `time.Parse(time.DateOnly, value)`.
