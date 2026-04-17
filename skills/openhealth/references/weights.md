# Weight Task Recipes

Use `agentops.RunWeightTask` for routine local weight tasks. The API accepts
simple JSON-friendly request fields and returns deterministic JSON-friendly
results.

## Write, Reapply, Or Correct Weights

Use `WeightTaskActionUpsert` with one or more `WeightInput` values. Upsert is
idempotent for the same date and unit: it returns `created`, `already_exists`,
or `updated`.

```go
result, err := agentops.RunWeightTask(context.Background(), client.LocalConfig{}, agentops.WeightTaskRequest{
	Action: agentops.WeightTaskActionUpsert,
	Weights: []agentops.WeightInput{
		{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
	},
})
```

## Latest And History

For only the latest entry:

```go
result, err := agentops.RunWeightTask(context.Background(), client.LocalConfig{}, agentops.WeightTaskRequest{
	Action:   agentops.WeightTaskActionList,
	ListMode: agentops.WeightListModeLatest,
})
```

For history, optionally set `Limit`. Use history mode for "two most recent" or
any count greater than one; `WeightListModeLatest` always returns one row.

```go
result, err := agentops.RunWeightTask(context.Background(), client.LocalConfig{}, agentops.WeightTaskRequest{
	Action:   agentops.WeightTaskActionList,
	ListMode: agentops.WeightListModeHistory,
	Limit:    2,
})
```

## Bounded History

For bounded date ranges, use strict date-only inclusive bounds:

```go
result, err := agentops.RunWeightTask(context.Background(), client.LocalConfig{}, agentops.WeightTaskRequest{
	Action:   agentops.WeightTaskActionList,
	ListMode: agentops.WeightListModeRange,
	FromDate: "2026-03-29",
	ToDate:   "2026-03-30",
})
```

When reporting the result, mirror every row in the JSON `entries` array, newest
first. AgentOps `entries` are already newest-first; do not inspect
implementation details to verify ordering. Do not report only the latest row,
and do not include older or newer records that were not part of the requested
range. Do not mention excluded dates at all, even to say they were excluded.

## Validation

Reject directly without running code when:

- a short date like `03/29` has no explicit year context,
- a year-first slash date is provided, such as `2026/03/31`,
- a date has a year but cannot be converted to `YYYY-MM-DD`,
- a value is non-positive or missing,
- a unit is not `lb`, `lbs`, `pound`, or `pounds`.

Explicit month/day/year dates with a year, such as `03/29/2026`, may be
converted to `YYYY-MM-DD` before calling AgentOps.

When the request is valid, the AgentOps facade performs the same validation
before opening the local database and returns `Rejected: true` with
`rejection_reason` for invalid task input.
