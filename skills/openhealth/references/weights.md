# Weight Task Recipes

Use `agentops.RunWeightTask` for routine local weight tasks. It returns
JSON-friendly write statuses, newest-first entries, and rejection reasons.
This reference is the task contract for routine agent use; do not inspect source
or test files to rediscover these shapes unless a task run fails.

## Request And Result Fields

```go
agentops.WeightTaskRequest{
	Action:   string,
	Weights:  []agentops.WeightInput,
	ListMode: string,
	FromDate: string,
	ToDate:   string,
	Limit:    int,
}

agentops.WeightInput{Date: string, Value: float64, Unit: string}
```

`WeightTaskResult` encodes to JSON with `rejected`, `rejection_reason`, `writes`,
`entries`, and `summary`. Each write has `date`, `value`, `unit`, and `status`.
Each entry has `date`, `value`, and `unit`.

## Write, Reapply, Or Correct Weights

Use `WeightTaskActionUpsert` with one or more `WeightInput` values. Repeating a
same-date value is idempotent, and a same-date different value updates the
existing row.

```go
result, err := agentops.RunWeightTask(context.Background(), client.LocalConfig{}, agentops.WeightTaskRequest{
	Action: agentops.WeightTaskActionUpsert,
	Weights: []agentops.WeightInput{
		{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
	},
})
```

Accepted units are `lb`, `lbs`, `pound`, and `pounds`; AgentOps normalizes them
to `lb`. For same-date corrections, one upsert request with the corrected value
is enough; the result `writes` status is `updated` and `entries` contains the
stored newest-first rows.

## Read Weights

```go
// latest only
agentops.WeightTaskRequest{
	Action:   agentops.WeightTaskActionList,
	ListMode: agentops.WeightListModeLatest,
}

// history, optionally limited
agentops.WeightTaskRequest{
	Action:   agentops.WeightTaskActionList,
	ListMode: agentops.WeightListModeHistory,
	Limit:    2,
}

// inclusive bounded date range
agentops.WeightTaskRequest{
	Action:   agentops.WeightTaskActionList,
	ListMode: agentops.WeightListModeRange,
	FromDate: "2026-03-29",
	ToDate:   "2026-03-30",
}
```

For "two most recent" or any count greater than one, use
`WeightListModeHistory` with `Limit`; `WeightListModeLatest` returns one row.

## Validation

Reject without writing when a request has an ambiguous short date, year-first
slash date, non-positive or missing value, or unsupported unit. Valid requests
are also validated by AgentOps before database access.
