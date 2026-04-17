---
name: agentops-code
description: Use this skill when evaluating OpenHealth weight tasks through the code-first AgentOps facade.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and the openhealth repository checkout.
---

# OpenHealth AgentOps Code Facade

Use the code-first `agentops` facade for local OpenHealth weight tasks. Do not
use the `openhealth` CLI. Do not inspect generated files, the Go module cache,
or SQLite directly. Do not run `bd prime`, repo-wide searches, or file listings.
The task API below is complete for this eval.

Run from the repository root. Create any temporary Go module outside the
repository with `mktemp -d`, use a `replace` directive back to the repository,
print JSON, then answer only from that JSON.

## One-Command Runner Pattern

Use one shell command shaped like this, changing only the request literal. Keep
`context.Background()` and `GOPROXY=off GOSUMDB=off go run -mod=mod .` exactly
as shown so the command stays offline and does not need retries:

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

## Requests

For add, reapply, or same-date correction tasks, use
`Action: agentops.WeightTaskActionUpsert` with one or more `WeightInput` values.
The result includes write statuses and newest-first entries.

For latest/history tasks:

```go
Action: agentops.WeightTaskActionList,
ListMode: agentops.WeightListModeHistory,
```

For bounded date ranges, use date-only inclusive bounds:

```go
Action: agentops.WeightTaskActionList,
ListMode: agentops.WeightListModeRange,
FromDate: "2026-03-29",
ToDate: "2026-03-30",
```

For invalid input or ambiguous short-date requests, do not run code. Reject the
request directly when the date is not `YYYY-MM-DD`, the value is non-positive,
or the unit is not `lb`, `lbs`, `pound`, or `pounds`.

When reporting results, convert JSON entries into plain rows like
`2026-03-30 151.6 lb`, newest first. For bounded range results, mirror every row
in the JSON `entries` array and do not mention excluded dates.
