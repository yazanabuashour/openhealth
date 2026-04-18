# openhealth

OpenHealth is a local-first health data runtime for agents. It ships a small
`openhealth` runner and a single-file skill.

## Quickstart

### Agent Install

#### Tell Your Agent

```text
Install https://github.com/yazanabuashour/openhealth
```

This installs the `openhealth` runner binary and the OpenHealth skill.

### Manual Install, Latest Release

```bash
curl -fsSL https://github.com/yazanabuashour/openhealth/releases/latest/download/install.sh | sh
```

Use this for quick local setup when you want the current release.

### Manual Install, Pinned Version

```bash
curl -fsSL https://github.com/yazanabuashour/openhealth/releases/download/v0.1.0/install.sh | sh
```

Use this for reproducible setup. Both manual install commands install the
matching `openhealth` runner binary, put it on `PATH` when possible, and then
install the same-tag skills.sh skill with:

```bash
npx -y skills add https://github.com/yazanabuashour/openhealth/tree/<tag>/skills/openhealth --skill openhealth --agent codex --global --yes
```

### Runner Interface

The skill calls these runner domains:

```bash
openhealth weight
openhealth blood-pressure
openhealth medications
openhealth labs
```

The runner reads structured JSON from stdin, validates and normalizes the
request, performs the local health operation, and writes structured JSON to
stdout.

## Local Go SDK

Go developers can embed the same local runtime directly:

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
same local health service used by the runner. There is no hosted service, remote
HTTP API, or generated API contract in the `0.1.0` release surface.

## Local Storage

By default, the local runtime stores its SQLite database under
`${XDG_DATA_HOME:-~/.local/share}/openhealth/openhealth.db`.

Override the default location with either:

- `client.LocalConfig{DataDir: "..."}`
- `client.LocalConfig{DatabasePath: "..."}`
- `OPENHEALTH_DATA_DIR`
- `OPENHEALTH_DATABASE_PATH`

The database path override wins over the data directory override.

## Eval Evidence

The production runner passed the release eval gate. In the CLI comparison run
documented in
[`docs/agent-eval-results/oh-5yr-maturity-throughput-final.md`](docs/agent-eval-results/oh-5yr-maturity-throughput-final.md),
the runner matched or improved correctness while using fewer tools, fewer
non-cached input tokens, and less wall time than CLI:

| Metric | Runner | CLI |
| --- | ---: | ---: |
| Tools | 28 | 57 |
| Non-cached input tokens | 105,247 | 121,774 |
| Wall time | 247.25s | 285.82s |

## Contributing and Maintainer Setup

Repository development uses the full local toolchain:

```bash
mise install
OPENHEALTH_DATA_DIR="$(mktemp -d)" mise exec -- go run ./examples/client_summary
mise exec -- go generate ./...
mise exec -- gofmt -w .
mise exec -- golangci-lint run
mise exec -- go test ./...
./scripts/validate-agent-skill.sh skills/openhealth
```

See `CONTRIBUTING.md` for contributor expectations and `docs/maintainers.md`
for maintainer-only workflow details.

## Repository Contents

- `CONTRIBUTING.md` explains how outside contributors should propose changes.
- `SECURITY.md` explains how to report vulnerabilities privately and what
  response timing to expect.
- `docs/release-verification.md` explains how to verify published source
  releases.
- `docs/agent-evals.md` explains how to evaluate production agent workflows.

## Release Contract

The `0.1.0` release deliverables are:

- platform archives for the `openhealth` binary
- the single-file `openhealth` skill archive
- the release installer script
- the Go module import path rooted at `github.com/yazanabuashour/openhealth`
- the direct-local Go package at `github.com/yazanabuashour/openhealth/client`

The release workflow is built around semantic version tags in the `v0.y.z`
range. Each tagged GitHub Release publishes binary archives, the skill archive,
a release installer, a canonical source archive, SHA256 checksums, an SPDX SBOM,
and GitHub attestations for release verification.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests.
Beads is maintainer-only workflow tooling and is not required for community
contributions.

See `CONTRIBUTING.md` for contribution expectations and `CODE_OF_CONDUCT.md`
for community standards.
