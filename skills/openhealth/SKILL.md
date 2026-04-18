---
name: openhealth
description: Use this skill when an agent needs to read or write local-first OpenHealth data through the AgentOps task facades in github.com/yazanabuashour/openhealth.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and the openhealth repository checkout.
---

# OpenHealth AgentOps

Use this skill for local-first OpenHealth user-data tasks. The production agent
interface is the `agentops` package. Use `agentops.RunWeightTask` for weight
tasks and `agentops.RunBloodPressureTask` for blood-pressure tasks. Supported
routine tasks are:

- local weight add, reapply, correction, latest, history, bounded-range, and validation requests; see [references/weights.md](references/weights.md)
- local blood-pressure record, correction, latest, history, bounded-range, and validation requests; see [references/blood-pressure.md](references/blood-pressure.md)

Use the quick recipes below for routine tasks. Open the linked references only
if a supported request needs a detail not covered here.

For supported tasks, pipe one JSON request into the matching AgentOps runner:
`go run ./cmd/openhealth-agentops weight` or
`go run ./cmd/openhealth-agentops blood-pressure`. The runner calls the matching
AgentOps task function with `client.LocalConfig{}` and prints JSON. Answer only
from that JSON.

For unsupported OpenHealth workflows, say the production AgentOps skill does not
support that workflow yet. Do not switch to a different interface unless the user
explicitly asks for one.

Do not write local OpenHealth data through SQLite directly. Do not inspect
generated API bindings or generated server code for routine user-data tasks.
Do not run repo-wide file discovery or broad searches for supported tasks. This
skill and its linked references are the routine task contract; do not inspect
source files, tests, generated code, or module-cache docs to rediscover
request/result shapes before the first task run. Only search the repository if
the AgentOps task run fails in a way that requires debugging the local checkout.

## Quick JSON Recipes

Weight actions are `upsert_weights`, `list_weights`, and `validate`. Weight
list modes are `latest`, `history`, and `range`.

```json
{"action":"upsert_weights","weights":[{"date":"2026-03-29","value":152.2,"unit":"lb"},{"date":"2026-03-30","value":151.6,"unit":"lb"}]}
{"action":"list_weights","list_mode":"latest"}
{"action":"list_weights","list_mode":"history","limit":2}
{"action":"list_weights","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

Blood-pressure actions are `record_blood_pressure`,
`correct_blood_pressure`, `list_blood_pressure`, and `validate`.
Blood-pressure list modes are `latest`, `history`, and `range`.

```json
{"action":"record_blood_pressure","readings":[{"date":"2026-03-29","systolic":122,"diastolic":78,"pulse":64},{"date":"2026-03-30","systolic":118,"diastolic":76}]}
{"action":"correct_blood_pressure","readings":[{"date":"2026-03-29","systolic":121,"diastolic":77}]}
{"action":"list_blood_pressure","list_mode":"latest"}
{"action":"list_blood_pressure","list_mode":"history","limit":2}
{"action":"list_blood_pressure","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

## Runner Pattern

Use this shape for supported tasks, changing only the JSON request and runner
domain:

```bash
go run ./cmd/openhealth-agentops weight <<'EOF'
{
  "action": "list_weights",
  "list_mode": "latest"
}
EOF
```

Use strict `YYYY-MM-DD` dates in AgentOps requests. If the user gives an
ambiguous short date without enough year context, ask for the year before
writing. If the user gives a year-first slash date such as `2026/03/31`, reject
it instead of rewriting it. Explicit month/day/year dates with a year, such as
`03/29/2026`, may be converted to `YYYY-MM-DD`.

For clearly invalid requests covered by this skill, reject directly without running code.
When reporting runner results, answer from JSON `entries`, `writes`, or
`rejection_reason`. AgentOps `entries` are already newest-first.
