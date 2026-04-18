# Lab Task Recipes

Use `agentops.RunLabTask` for routine local lab collection tasks. In the
production skill, send request JSON to `go run ./cmd/openhealth-agentops labs`.

## Request And Result Fields

Request JSON fields are `action`, `collections`, `collection`, `target`,
`list_mode`, `from_date`, `to_date`, `limit`, and `analyte_slug`. Each collection
has `date` and nested `panels`; each panel has `panel_name` and `results`; each
result has `test_name`, optional `canonical_slug`, `value_text`, optional
`value_numeric`, optional `units`, optional `range_text`, and optional `flag`.

Use these exact JSON values:

- `action`: `record_labs`, `correct_labs`, `delete_labs`, `list_labs`, or
  `validate`
- `list_mode`: `latest`, `history`, or `range`
- `canonical_slug` and `analyte_slug`: `tsh`, `free-t4`, `cholesterol-total`,
  `ldl`, `hdl`, `triglycerides`, or `glucose`

`LabTaskResult` encodes to JSON with `rejected`, `rejection_reason`, `writes`,
`entries`, and `summary`. Entries include collection `id`, `date`, nested
`panels`, and nested `results`.

## Record Labs

Use `record_labs` with one or more date-only collections:

```json
{
  "action": "record_labs",
  "collections": [
    {
      "date": "2026-03-29",
      "panels": [
        {
          "panel_name": "Metabolic",
          "results": [
            {
              "test_name": "Glucose",
              "canonical_slug": "glucose",
              "value_text": "89",
              "value_numeric": 89,
              "units": "mg/dL",
              "range_text": "70-99"
            }
          ]
        }
      ]
    }
  ]
}
```

Repeating an exact same-date collection is idempotent and returns
`already_exists`. A same-date collection with different panels or results is
rejected; use `correct_labs`.

## Correct Or Delete Labs

Use `correct_labs` or `delete_labs` with a target by `id`, or by `date`:

```json
{
  "action": "correct_labs",
  "target": {"date": "2026-03-29"},
  "collection": {
    "date": "2026-03-29",
    "panels": [
      {
        "panel_name": "Thyroid",
        "results": [
          {"test_name": "TSH", "canonical_slug": "tsh", "value_text": "3.1", "value_numeric": 3.1, "units": "uIU/mL"}
        ]
      }
    ]
  }
}
```

The target must match exactly one lab collection. If no collection or multiple
collections match, AgentOps returns `rejected` with a `rejection_reason`.

## Read Labs

```json
{"action":"list_labs","list_mode":"latest"}
{"action":"list_labs","list_mode":"history","limit":2}
{"action":"list_labs","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
{"action":"list_labs","list_mode":"latest","analyte_slug":"glucose"}
```

For "two most recent" or any count greater than one, use `history` with
`limit`; `latest` returns one matching collection. `analyte_slug` filters nested
results to that canonical analyte and omits collections without matching
results.

## Validation

Reject without writing when a request has an ambiguous short date, year-first
slash date, missing panels, missing results, missing `test_name`, missing
`value_text`, unsupported analyte slug, or empty optional text field. Valid
requests are also validated by AgentOps before database access.
