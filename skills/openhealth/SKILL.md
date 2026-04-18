---
name: openhealth
description: Use this skill for local-first OpenHealth weight, blood-pressure, medication, or lab data through AgentOps; for ambiguous short dates, year-first slash dates, invalid values, unsupported units/analytes/statuses, empty optional text fields, unsafe corrections/deletes, or medication end dates before start dates, reject directly without tools.
license: MIT
compatibility: Requires local filesystem access and an installed openhealth-agentops binary on PATH.
---

# OpenHealth AgentOps

Use `agentops.RunWeightTask`, `agentops.RunBloodPressureTask`,
`agentops.RunMedicationTask`, and `agentops.RunLabTask` through the JSON
runners:

- `openhealth-agentops weight`
- `openhealth-agentops blood-pressure`
- `openhealth-agentops medications`
- `openhealth-agentops labs`

## Reject Before Tools

For the cases below, reject directly without running code, inspecting files,
searching the repo, checking the database, using the AgentOps runner, or calling
the CLI when the request has:

| Issue | Response |
| --- | --- |
| ambiguous short date without year context, like `03/29` | ask for the year |
| year-first slash date, like `2026/03/31` | require `YYYY-MM-DD` |
| non-positive weight, systolic, diastolic, or pulse | reject as invalid |
| unsupported weight unit, like `stone` | reject; pounds only |
| unsupported lab analyte slug | reject as invalid |
| unsupported medication status | reject as invalid |
| empty optional medication/lab text field | reject as invalid |
| medication end date before start date | reject as invalid |

`03/29/2026` may be normalized to `2026-03-29`.

## Runner Contract

Pipe one JSON request to one runner and answer only from JSON `entries`,
`writes`, or `rejection_reason`. Run mixed requests as one call per domain.
AgentOps `entries` are already newest-first. Valid requests are validated by
AgentOps before database access.

Weights:

```json
{"action":"upsert_weights","weights":[{"date":"2026-03-29","value":152.2,"unit":"lb"}]}
{"action":"list_weights","list_mode":"latest"}
{"action":"list_weights","list_mode":"history","limit":2}
{"action":"list_weights","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

Request JSON fields are `action`, `weights`, `list_mode`, `from_date`,
`to_date`, and `limit`. Each weight has `date`, `value`, and `unit`.
Use `upsert_weights` to write, reapply, or correct weights. Repeating a
same-date value is idempotent. A same-date different value updates the row.
Accepted units are `lb`, `lbs`, `pound`, and `pounds`; AgentOps normalizes them
to `lb`. For "two most recent" or any count greater than one, use `history`
with `limit`; `latest` returns one row.

Blood pressure:

```json
{"action":"record_blood_pressure","readings":[{"date":"2026-03-29","systolic":122,"diastolic":78,"pulse":64}]}
{"action":"correct_blood_pressure","readings":[{"date":"2026-03-29","systolic":121,"diastolic":77}]}
{"action":"list_blood_pressure","list_mode":"latest"}
{"action":"list_blood_pressure","list_mode":"history","limit":2}
{"action":"list_blood_pressure","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

Request JSON fields are `action`, `readings`, `list_mode`, `from_date`,
`to_date`, and `limit`. Each reading has `date`, `systolic`, `diastolic`, and
optional positive integer `pulse`. Use `record_blood_pressure` to add readings.
Use `correct_blood_pressure` when the user asks to correct an existing
same-date reading. Correction updates exactly one existing reading on the
requested date; if no same-date reading or multiple same-date readings exist,
AgentOps returns `rejected` with `rejection_reason`. For "two most recent" or
any count greater than one, use `history` with `limit`; `latest` returns one row.

Medications:

```json
{"action":"record_medications","medications":[{"name":"Levothyroxine","dosage_text":"25 mcg","start_date":"2026-01-01"}]}
{"action":"correct_medication","target":{"name":"Levothyroxine","start_date":"2026-01-01"},"medication":{"name":"Levothyroxine","dosage_text":"50 mcg","start_date":"2026-01-01","end_date":"2026-04-01"}}
{"action":"delete_medication","target":{"name":"Levothyroxine","start_date":"2026-01-01"}}
{"action":"list_medications","status":"active"}
{"action":"list_medications","status":"all"}
```

Request JSON fields are `action`, `medications`, `medication`, `target`, and
`status`. Each medication request has `name`, optional `dosage_text`,
`start_date`, and optional `end_date`. Use `record_medications` with one or more
courses. Repeating an exact same `name` and `start_date` course is idempotent
and returns `already_exists`; the same `name` and `start_date` with different
details is rejected and should be corrected with `correct_medication`.
Use `correct_medication` or `delete_medication` with a target by `id`, or by
exact normalized `name` and `start_date`. The target must match exactly one
medication; zero or multiple matches return `rejected` with `rejection_reason`.
`active` is the default status; only `active` and `all` are supported.

Labs:

```json
{"action":"record_labs","collections":[{"date":"2026-03-29","panels":[{"panel_name":"Metabolic","results":[{"test_name":"Glucose","canonical_slug":"glucose","value_text":"89","value_numeric":89,"units":"mg/dL","range_text":"70-99"}]}]}]}
{"action":"correct_labs","target":{"date":"2026-03-29"},"collection":{"date":"2026-03-29","panels":[{"panel_name":"Thyroid","results":[{"test_name":"TSH","canonical_slug":"tsh","value_text":"3.1","value_numeric":3.1,"units":"uIU/mL"}]}]}}
{"action":"delete_labs","target":{"date":"2026-03-29"}}
{"action":"list_labs","list_mode":"latest"}
{"action":"list_labs","list_mode":"history","limit":2}
{"action":"list_labs","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
{"action":"list_labs","list_mode":"latest","analyte_slug":"glucose"}
```

Request JSON fields are `action`, `collections`, `collection`, `target`,
`list_mode`, `from_date`, `to_date`, `limit`, and `analyte_slug`. Each
collection request has `date`, nested `panels`, and nested `results`. Each panel
has `panel_name` and `results`. Each result has `test_name`, optional
`canonical_slug`, `value_text`, optional `value_numeric`, optional `units`,
optional `range_text`, and optional `flag`.

Supported `canonical_slug` and `analyte_slug` values are `tsh`, `free-t4`,
`cholesterol-total`, `ldl`, `hdl`, `triglycerides`, and `glucose`.
Use `record_labs` with one or more date-only collections. Repeating an exact
same-date collection is idempotent and returns `already_exists`; a same-date
collection with different panels or results is rejected and should be corrected
with `correct_labs`. Use `correct_labs` or `delete_labs` with a target by `id`,
or by `date`. The target must match exactly one lab collection; zero or multiple
matches return `rejected` with `rejection_reason`.

For "two most recent" or any count greater than one, use `history` with
`limit`; `latest` returns one matching collection. `analyte_slug` filters nested
results to that canonical analyte and omits collections without matching
results.

Do not run repo-wide file discovery or broad searches for routine user-data
tasks; do not inspect generated files, module-cache docs, or SQLite directly.
