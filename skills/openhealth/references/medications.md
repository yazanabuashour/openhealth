# Medication Task Recipes

Use `agentops.RunMedicationTask` for routine local medication-course tasks. In
the production skill, send request JSON to
`go run ./cmd/openhealth-agentops medications`.

## Request And Result Fields

Request JSON fields are `action`, `medications`, `medication`, `target`, and
`status`. Each medication has `name`, optional `dosage_text`, `start_date`, and
optional `end_date`.

Use these exact JSON values:

- `action`: `record_medications`, `correct_medication`, `delete_medication`,
  `list_medications`, or `validate`
- `status`: `active` or `all`

`MedicationTaskResult` encodes to JSON with `rejected`, `rejection_reason`,
`writes`, `entries`, and `summary`. Each write and entry has `id`, `name`,
optional `dosage_text`, `start_date`, and optional `end_date`. Writes also have
`status`.

## Record Medications

Use `record_medications` with one or more medications:

```json
{
  "action": "record_medications",
  "medications": [
    {"name": "Levothyroxine", "dosage_text": "25 mcg", "start_date": "2026-01-01"}
  ]
}
```

Repeating an exact same `name` and `start_date` course is idempotent and returns
`already_exists`. A same `name` and `start_date` with different details is
rejected; use `correct_medication`.

## Correct Or Delete Medications

Use `correct_medication` or `delete_medication` with a target by `id`, or by
exact `name` and `start_date`:

```json
{
  "action": "correct_medication",
  "target": {"name": "Levothyroxine", "start_date": "2026-01-01"},
  "medication": {"name": "Levothyroxine", "dosage_text": "50 mcg", "start_date": "2026-01-01"}
}
```

The target must match exactly one medication. If no medication or multiple
medications match, AgentOps returns `rejected` with a `rejection_reason`.

## Read Medications

```json
{"action":"list_medications","status":"active"}
{"action":"list_medications","status":"all"}
```

`active` is the default status.

## Validation

Reject without writing when a request has an ambiguous short date, year-first
slash date, missing name, missing or malformed `start_date`, empty optional text
field, unsupported status, or `end_date` before `start_date`. Valid requests are
also validated by AgentOps before database access.
