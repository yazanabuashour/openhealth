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

For supported tasks, create a temporary Go module outside the repository, replace
`github.com/yazanabuashour/openhealth` to the current checkout, call the matching
AgentOps task function with `client.LocalConfig{}`, print JSON, and answer only
from that JSON.

For unsupported OpenHealth workflows, say the production AgentOps skill does not
support that workflow yet. Do not switch to a different interface unless the user
explicitly asks for one.

Do not write local OpenHealth data through SQLite directly. Do not inspect
generated API bindings or generated server code for routine user-data tasks.
The linked references are the routine task contract; do not inspect source files,
tests, generated code, or module-cache docs to rediscover request/result shapes
before the first task run. Only search the repository if the AgentOps task run
fails in a way that requires debugging the local checkout.

## Runner Pattern

Use this shape for supported tasks, changing only the request literal and the
AgentOps function:

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
		Action: agentops.WeightTaskActionList,
		ListMode: agentops.WeightListModeLatest,
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

Use strict `YYYY-MM-DD` dates in AgentOps requests. If the user gives an
ambiguous short date without enough year context, ask for the year before
writing. If the user gives a year-first slash date such as `2026/03/31`, reject
it instead of rewriting it. Explicit month/day/year dates with a year, such as
`03/29/2026`, may be converted to `YYYY-MM-DD`.

When reporting results, answer from JSON `entries`, `writes`, or
`rejection_reason`. AgentOps `entries` are already newest-first.
