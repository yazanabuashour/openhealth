# Blood Pressure Task Recipes

Use `agentops.RunBloodPressureTask` for routine local blood-pressure tasks. It
returns JSON-friendly write statuses, newest-first entries, and rejection
reasons.
This reference is the task contract for routine agent use; do not inspect source
or test files to rediscover these shapes unless a task run fails.

## Request And Result Fields

```go
agentops.BloodPressureTaskRequest{
	Action:   string,
	Readings: []agentops.BloodPressureInput,
	ListMode: string,
	FromDate: string,
	ToDate:   string,
	Limit:    int,
}

agentops.BloodPressureInput{
	Date:      string,
	Systolic:  int,
	Diastolic: int,
	Pulse:     *int,
}
```

`BloodPressureTaskResult` encodes to JSON with `rejected`, `rejection_reason`,
`writes`, `entries`, and `summary`. Each write and entry has `date`, `systolic`,
`diastolic`, and optional `pulse`. Record and correction results also return
`entries` newest-first after the write.

## Record Blood Pressure

Use `BloodPressureTaskActionRecord` with one or more readings:

```go
result, err := agentops.RunBloodPressureTask(context.Background(), client.LocalConfig{}, agentops.BloodPressureTaskRequest{
	Action: agentops.BloodPressureTaskActionRecord,
	Readings: []agentops.BloodPressureInput{
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
	},
})
```

Pulse is optional. When present, pass it as a positive integer pointer:

```go
pulse := 64
result, err := agentops.RunBloodPressureTask(context.Background(), client.LocalConfig{}, agentops.BloodPressureTaskRequest{
	Action: agentops.BloodPressureTaskActionRecord,
	Readings: []agentops.BloodPressureInput{
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: &pulse},
	},
})
```

## Correct Blood Pressure

Use `BloodPressureTaskActionCorrect` when the user asks to correct an existing
same-date reading:

```go
result, err := agentops.RunBloodPressureTask(context.Background(), client.LocalConfig{}, agentops.BloodPressureTaskRequest{
	Action: agentops.BloodPressureTaskActionCorrect,
	Readings: []agentops.BloodPressureInput{
		{Date: "2026-03-29", Systolic: 121, Diastolic: 77},
	},
})
```

Correction updates exactly one existing reading on the requested date. If there
is no same-date reading, or there are multiple same-date readings, AgentOps
returns `rejected` with a `rejection_reason` instead of guessing.

## Read Blood Pressure

```go
// latest only
agentops.BloodPressureTaskRequest{
	Action:   agentops.BloodPressureTaskActionList,
	ListMode: agentops.BloodPressureListModeLatest,
}

// history, optionally limited
agentops.BloodPressureTaskRequest{
	Action:   agentops.BloodPressureTaskActionList,
	ListMode: agentops.BloodPressureListModeHistory,
	Limit:    2,
}

// inclusive bounded date range
agentops.BloodPressureTaskRequest{
	Action:   agentops.BloodPressureTaskActionList,
	ListMode: agentops.BloodPressureListModeRange,
	FromDate: "2026-03-29",
	ToDate:   "2026-03-30",
}
```

For "two most recent" or any count greater than one, use
`BloodPressureListModeHistory` with `Limit`; `BloodPressureListModeLatest`
returns one row.

## Validation

Reject without writing when a request has an ambiguous short date, year-first
slash date, non-positive systolic, non-positive diastolic, or non-positive pulse.
Valid requests are also validated by AgentOps before database access.
