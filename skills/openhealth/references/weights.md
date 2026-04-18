# Weight Task Recipes

Use `agentops.RunWeightTask` for routine local weight tasks. It returns
JSON-friendly write statuses, newest-first entries, and rejection reasons.
This reference is the task contract for routine agent use; do not inspect source
or test files to rediscover these shapes unless a task run fails. In the
production skill, send this request JSON to
`go run ./cmd/openhealth-agentops weight`.

## Request And Result Fields

Request JSON fields are `action`, `weights`, `list_mode`, `from_date`,
`to_date`, and `limit`. Each weight has `date`, `value`, and `unit`.

Use these exact JSON values:

- `action`: `upsert_weights`, `list_weights`, or `validate`
- `list_mode`: `latest`, `history`, or `range`

`WeightTaskResult` encodes to JSON with `rejected`, `rejection_reason`, `writes`,
`entries`, and `summary`. Each write has `date`, `value`, `unit`, and `status`.
Each entry has `date`, `value`, and `unit`.

## Write, Reapply, Or Correct Weights

Use `upsert_weights` with one or more weights. Repeating a same-date value is
idempotent, and a same-date different value updates the existing row.

```json
{
  "action": "upsert_weights",
  "weights": [
    {"date": "2026-03-29", "value": 152.2, "unit": "lb"},
    {"date": "2026-03-30", "value": 151.6, "unit": "lb"}
  ]
}
```

Accepted units are `lb`, `lbs`, `pound`, and `pounds`; AgentOps normalizes them
to `lb`. For same-date corrections, one upsert request with the corrected value
is enough; the result `writes` status is `updated` and `entries` contains the
stored newest-first rows.

## Read Weights

```json
{"action":"list_weights","list_mode":"latest"}
{"action":"list_weights","list_mode":"history","limit":2}
{"action":"list_weights","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

For "two most recent" or any count greater than one, use
`WeightListModeHistory` with `Limit`; `WeightListModeLatest` returns one row.

## Validation

Reject without writing when a request has an ambiguous short date, year-first
slash date, non-positive or missing value, or unsupported unit. Valid requests
are also validated by AgentOps before database access.
