---
name: openhealth
description: Use this skill when an agent needs to read or write local-first OpenHealth data through the AgentOps task facade and Go SDK in github.com/yazanabuashour/openhealth.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and the openhealth repository checkout.
---

# OpenHealth AgentOps

Use this skill for local-first OpenHealth tasks. For routine local weight tasks,
the production agent path is the task-shaped `agentops` facade, not the
`openhealth` CLI and not direct SQLite access.

## Default Path

- Run from the repository root.
- For routine weight add, reapply, correction, latest, history, or bounded-range
  requests, use [references/weights.md](references/weights.md) and
  `github.com/yazanabuashour/openhealth/agentops`.
- `agentops.RunWeightTask(context.Background(), client.LocalConfig{}, request)`
  honors the configured local environment, including
  `OPENHEALTH_DATABASE_PATH`. Do not search for the database path unless the
  AgentOps run fails with a path/configuration error.
- The instructions on this page are complete for routine weight tasks. Do not
  run `bd prime`, `rg --files`, repo-wide `rg`, `find .`, `env`, `pwd`, or
  exploratory file listings before using AgentOps.
- The import paths, constants, request fields, module version, and command
  template below are exact. Do not read `go.mod`, search for `RunWeightTask`, or
  inspect source files to confirm them.
- Do not inspect `.agents` files after this skill is loaded. Do not inspect
  `references/weights.md` unless the request examples below are insufficient.
- Do not use `go run ./cmd/openhealth` for routine agent weight tasks unless the
  user explicitly asks for the CLI.
- Do not inspect `client.gen.go`, generated server code, the Go module cache,
  large dependency directories, or SQLite directly for routine weight tasks.
- For non-weight OpenHealth workflows, use the hand-written Go SDK facade in
  `github.com/yazanabuashour/openhealth/client`.

## Routine Weight Fast Path

For valid weight write/list requests, run one temporary Go program outside the
repository. Change only the request literal. Do not run separate setup,
environment, path, search, source-inspection, or discovery commands. Use the
single command below; `repo="$(pwd)"` is the only repository path lookup needed.

```bash
tmp="$(mktemp -d)" && repo="$(pwd)" && cat > "$tmp/go.mod" <<EOF
module openhealth-agentops-task

go 1.26.2

require github.com/yazanabuashour/openhealth v0.0.0

replace github.com/yazanabuashour/openhealth => $repo
EOF
cat > "$tmp/main.go" <<'EOF'
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/yazanabuashour/openhealth/agentops"
	"github.com/yazanabuashour/openhealth/client"
)

func main() {
	result, err := agentops.RunWeightTask(context.Background(), client.LocalConfig{}, agentops.WeightTaskRequest{
		Action: agentops.WeightTaskActionUpsert,
		Weights: []agentops.WeightInput{
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		log.Fatal(err)
	}
}
EOF
(cd "$tmp" && GOPROXY=off GOSUMDB=off go run -mod=mod .)
```

Use these request shapes:

```go
// add, reapply, or correction
agentops.WeightTaskRequest{
	Action: agentops.WeightTaskActionUpsert,
	Weights: []agentops.WeightInput{{Date: "2026-03-29", Value: 152.2, Unit: "lb"}},
}

// latest only
agentops.WeightTaskRequest{
	Action:   agentops.WeightTaskActionList,
	ListMode: agentops.WeightListModeLatest,
}

// history, optionally limited; use this for "two most recent"
agentops.WeightTaskRequest{
	Action:   agentops.WeightTaskActionList,
	ListMode: agentops.WeightListModeHistory,
	Limit:    2,
}

// bounded date range
agentops.WeightTaskRequest{
	Action:   agentops.WeightTaskActionList,
	ListMode: agentops.WeightListModeRange,
	FromDate: "2026-03-29",
	ToDate:   "2026-03-30",
}
```

Do not combine `WeightListModeLatest` with `Limit`. Latest mode returns one row.
For "two most recent" or any count greater than one, use
`WeightListModeHistory` with `Limit`.

For invalid input or ambiguous short-date requests, reject directly without
running code:

- Do not infer a year for short dates unless the user or conversation gives
  explicit year context.
- Do not write non-positive, missing, or otherwise invalid values.
- Do not convert unsupported units. Accepted units are `lb`, `lbs`, `pound`,
  and `pounds`, normalized to `lb`.
- Do not write year-first slash dates, such as `2026/03/31`; reject them
  without normalizing or rewriting them. Explicit month/day/year dates with a
  year, such as `03/29/2026`, may be converted to `YYYY-MM-DD`.

When reporting results, answer from the JSON `entries`, `writes`, or
`rejection_reason` fields only. AgentOps `entries` are already newest-first.
Convert entries into plain rows like `2026-03-30 151.6 lb`, newest first. For
bounded ranges, mirror every JSON entry from the requested range and do not
mention excluded dates.

## Generated Client Fallback

The generated OpenAPI client remains available for advanced API-contract work,
HTTP-server calls, or endpoints not yet covered by the SDK facade. Do not start
there for common agent tasks.
