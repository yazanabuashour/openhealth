# OpenHealth

OpenHealth is a local-first health data runtime for agents. The supported
agent path is a small `openhealth` runner plus a single-file skill.

## Install

Tell your agent:

```text
Install OpenHealth from https://github.com/yazanabuashour/openhealth.
Complete both required steps before reporting success:
1. Install and verify the openhealth runner binary with `openhealth --version`.
2. Register the OpenHealth skill from skills/openhealth/SKILL.md using your native skill system.
```

For the latest release:

```bash
sh -c "$(curl -fsSL https://github.com/yazanabuashour/openhealth/releases/latest/download/install.sh)"
```

For a pinned release:

```bash
OPENHEALTH_VERSION=v0.3.1 sh -c "$(curl -fsSL https://github.com/yazanabuashour/openhealth/releases/download/v0.3.1/install.sh)"
```

A complete install has two parts:

- `openhealth --version` succeeds
- the matching skill is registered from `skills/openhealth/SKILL.md`,
  `https://github.com/yazanabuashour/openhealth/tree/<tag>/skills/openhealth`,
  or `openhealth_<version>_skill.tar.gz`

Use the agent's native skill manager. OpenHealth does not require a specific
skill path or agent implementation.

## Upgrade

Tell your agent:

```text
Upgrade OpenHealth from https://github.com/yazanabuashour/openhealth.
Complete both required steps before reporting success:
1. Upgrade and verify the openhealth runner binary with `openhealth --version`.
2. Re-register the OpenHealth skill from skills/openhealth/SKILL.md using your native skill system.
```

Or upgrade the runner manually:

```bash
sh -c "$(curl -fsSL https://github.com/yazanabuashour/openhealth/releases/latest/download/install.sh)"
```

Then verify the runner and re-register the matching skill:

```bash
command -v openhealth
openhealth --version

```

## AgentOps Architecture

OpenHealth's agent-facing path is the AgentOps pattern: the skill gives the
agent task policy, and the local runner performs stateful health-data operations
through structured JSON. This keeps domain rules close to the agent, avoids
broad repo search and ad hoc human CLI flows, and leaves storage local instead
of requiring a hosted service.

OpenHealth treats this runner/skill architecture as its competitive interface
for agents compared with traditional MCP or CLI-only integrations. The current
evals show the production runner/skill matched or improved correctness while
using fewer tools, fewer non-cached input tokens, and less wall time than CLI.

## Runner Interface

The skill sends structured JSON on stdin and reads structured JSON from stdout
for these runner domains:

```bash
openhealth weight
openhealth body-composition
openhealth blood-pressure
openhealth medications
openhealth labs
openhealth imaging
openhealth sleep
```

## Direct Go Package

Go developers can import the same local runtime from the source module:

```bash
go get github.com/yazanabuashour/openhealth@main
```

Minimal usage from Go:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yazanabuashour/openhealth/client"
)

func main() {
	api, err := client.OpenLocal(client.LocalConfig{})
	if err != nil {
		log.Fatal(err)
	}
	defer api.Close()

	recordedAt, err := time.Parse(time.DateOnly, "2026-03-29")
	if err != nil {
		log.Fatal(err)
	}

	result, err := api.UpsertWeight(context.Background(), client.WeightRecordInput{
		RecordedAt: recordedAt,
		Value:      152.2,
		Unit:       client.WeightUnitLb,
	})
	if err != nil {
		log.Fatal(err)
	}

	summary, err := api.Summary(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s %.1f lb %s\n", result.Entry.RecordedAt.Format(time.DateOnly), result.Entry.Value, result.Status)
	fmt.Printf("active medications=%d\n", summary.ActiveMedicationCount)
}
```

`client.OpenLocal(...)` opens SQLite locally, runs migrations, and calls the
same local health service used by the runner. This is a developer/source import
path for direct local embedding, not the agent installation path. OpenHealth
`0.3.1` does not ship a hosted service or remote HTTP API contract.

## Local Storage

The default SQLite path is
`${XDG_DATA_HOME:-~/.local/share}/openhealth/openhealth.db`. Override it with:

- `client.LocalConfig{DataDir: "..."}`
- `client.LocalConfig{DatabasePath: "..."}`
- `OPENHEALTH_DATA_DIR`
- `OPENHEALTH_DATABASE_PATH`

The database path override wins over the data directory override.

## Eval Evidence

The production runner/skill passed the latest 50-scenario release gate:
[`docs/agent-eval-results/oh-5yr-2026-04-20-v0.3.1-final.md`](docs/agent-eval-results/oh-5yr-2026-04-20-v0.3.1-final.md).
The CLI comparison found matching or improved correctness with fewer tools,
fewer non-cached input tokens, and less wall time:
[`docs/agent-eval-results/oh-5yr-maturity-throughput-final.md`](docs/agent-eval-results/oh-5yr-maturity-throughput-final.md).

The eval protocol is documented in [`docs/agent-evals.md`](docs/agent-evals.md).

## Development

Use the full local toolchain for repository development:

```bash
mise install
printf '%s\n' '{"action":"list_weights","list_mode":"latest"}' | \
  OPENHEALTH_DATA_DIR="$(mktemp -d)" mise exec -- go run ./cmd/openhealth weight
mise exec -- go generate ./...
mise exec -- gofmt -w .
mise exec -- golangci-lint run
mise exec -- go test ./...
./scripts/validate-agent-skill.sh skills/openhealth
```

`golangci-lint` is pinned by `mise.toml`; use `mise exec -- golangci-lint run`
for local checks.

## Releases

Tagged `v0.y.z` releases publish platform binary archives, the skill archive,
the installer, source archive, SHA256 checksums, an SPDX SBOM, and GitHub
attestations. Published release assets are intended to be immutable going
forward. See
[`docs/release-verification.md`](docs/release-verification.md) for verification
steps.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests.
Beads is maintainer-only workflow tooling and is not required for community
contributions.

See `CONTRIBUTING.md` for contribution expectations, `CODE_OF_CONDUCT.md` for
community standards, `SECURITY.md` for vulnerability reporting, and
`docs/maintainers.md` for maintainer-only workflow details.
