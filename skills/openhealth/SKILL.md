---
name: openhealth
description: Use this skill for local-first OpenHealth weight, blood-pressure, medication, or lab data through AgentOps; for ambiguous short dates, year-first slash dates, invalid values, unsupported units/analytes/statuses, or unsafe corrections/deletes, reject directly without tools.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and the openhealth repository checkout.
---

# OpenHealth AgentOps

Use `agentops.RunWeightTask`, `agentops.RunBloodPressureTask`,
`agentops.RunMedicationTask`, and `agentops.RunLabTask` through the JSON
runners:

- `go run ./cmd/openhealth-agentops weight`
- `go run ./cmd/openhealth-agentops blood-pressure`
- `go run ./cmd/openhealth-agentops medications`
- `go run ./cmd/openhealth-agentops labs`

## Reject Before Tools

For the cases below, reject directly without running code, opening references,
inspecting files, searching the repo, checking the database, using the AgentOps
runner, or calling the CLI when the request has:

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
AgentOps `entries` are already newest-first.

Weights:

```json
{"action":"upsert_weights","weights":[{"date":"2026-03-29","value":152.2,"unit":"lb"}]}
{"action":"list_weights","list_mode":"latest"}
{"action":"list_weights","list_mode":"history","limit":2}
{"action":"list_weights","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

Blood pressure:

```json
{"action":"record_blood_pressure","readings":[{"date":"2026-03-29","systolic":122,"diastolic":78,"pulse":64}]}
{"action":"correct_blood_pressure","readings":[{"date":"2026-03-29","systolic":121,"diastolic":77}]}
{"action":"list_blood_pressure","list_mode":"latest"}
{"action":"list_blood_pressure","list_mode":"history","limit":2}
{"action":"list_blood_pressure","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

Medications:

```json
{"action":"record_medications","medications":[{"name":"Levothyroxine","dosage_text":"25 mcg","start_date":"2026-01-01"}]}
{"action":"correct_medication","target":{"name":"Levothyroxine","start_date":"2026-01-01"},"medication":{"name":"Levothyroxine","dosage_text":"50 mcg","start_date":"2026-01-01","end_date":"2026-04-01"}}
{"action":"delete_medication","target":{"name":"Levothyroxine","start_date":"2026-01-01"}}
{"action":"list_medications","status":"active"}
{"action":"list_medications","status":"all"}
```

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

Open [references/weights.md](references/weights.md) or
[references/blood-pressure.md](references/blood-pressure.md),
[references/medications.md](references/medications.md), or
[references/labs.md](references/labs.md) only if a supported request needs
detail not shown here. Do not run repo-wide file discovery or broad searches for
routine user-data tasks; do not inspect generated files, module-cache docs, or
SQLite directly.
