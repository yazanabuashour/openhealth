---
name: openhealth
description: Use this skill for local-first OpenHealth weight or blood-pressure data through AgentOps; for ambiguous short dates, year-first slash dates, non-positive values, or unsupported weight units, reject directly without tools.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and the openhealth repository checkout.
---

# OpenHealth AgentOps

Use `agentops.RunWeightTask` and `agentops.RunBloodPressureTask` through the
JSON runners:

- `go run ./cmd/openhealth-agentops weight`
- `go run ./cmd/openhealth-agentops blood-pressure`

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

`03/29/2026` may be normalized to `2026-03-29`.

## Runner Contract

Pipe one JSON request to one runner and answer only from JSON `entries`,
`writes`, or `rejection_reason`. Run mixed weight and blood-pressure requests as
one call per domain. AgentOps `entries` are already newest-first.

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

Open [references/weights.md](references/weights.md) or
[references/blood-pressure.md](references/blood-pressure.md) only if a supported
request needs detail not shown here. Do not run repo-wide file discovery or broad searches
for routine user-data tasks; do not inspect generated files,
module-cache docs, or SQLite directly.
