# Blood Pressure Task Recipes

Use `agentops.RunBloodPressureTask` for routine local blood-pressure tasks. It
returns JSON-friendly write statuses, newest-first entries, and rejection
reasons.
This reference is the task contract for routine agent use; do not inspect source
or test files to rediscover these shapes unless a task run fails. In the
production skill, send this request JSON to
`go run ./cmd/openhealth-agentops blood-pressure`.

## Request And Result Fields

Request JSON fields are `action`, `readings`, `list_mode`, `from_date`,
`to_date`, and `limit`. Each reading has `date`, `systolic`, `diastolic`, and
optional `pulse`.

Use these exact JSON values:

- `action`: `record_blood_pressure`, `correct_blood_pressure`,
  `list_blood_pressure`, or `validate`
- `list_mode`: `latest`, `history`, or `range`

`BloodPressureTaskResult` encodes to JSON with `rejected`, `rejection_reason`,
`writes`, `entries`, and `summary`. Each write and entry has `date`, `systolic`,
`diastolic`, and optional `pulse`. Record and correction results also return
`entries` newest-first after the write.

## Record Blood Pressure

Use `record_blood_pressure` with one or more readings:

```json
{
  "action": "record_blood_pressure",
  "readings": [
    {"date": "2026-03-29", "systolic": 122, "diastolic": 78},
    {"date": "2026-03-30", "systolic": 118, "diastolic": 76}
  ]
}
```

Pulse is optional. When present, pass it as a positive integer:

```json
{
  "action": "record_blood_pressure",
  "readings": [
    {"date": "2026-03-29", "systolic": 122, "diastolic": 78, "pulse": 64}
  ]
}
```

## Correct Blood Pressure

Use `correct_blood_pressure` when the user asks to correct an existing same-date
reading:

```json
{
  "action": "correct_blood_pressure",
  "readings": [
    {"date": "2026-03-29", "systolic": 121, "diastolic": 77}
  ]
}
```

Correction updates exactly one existing reading on the requested date. If there
is no same-date reading, or there are multiple same-date readings, AgentOps
returns `rejected` with a `rejection_reason` instead of guessing.

## Read Blood Pressure

```json
{"action":"list_blood_pressure","list_mode":"latest"}
{"action":"list_blood_pressure","list_mode":"history","limit":2}
{"action":"list_blood_pressure","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

For "two most recent" or any count greater than one, use
`BloodPressureListModeHistory` with `Limit`; `BloodPressureListModeLatest`
returns one row.

## Validation

Reject without writing when a request has an ambiguous short date, year-first
slash date, non-positive systolic, non-positive diastolic, or non-positive pulse.
Valid requests are also validated by AgentOps before database access.
