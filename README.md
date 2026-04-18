# openhealth

OpenHealth is a local-first health runtime for agent-operated personal health
data. The production agent surface is the `openhealth` binary plus the
single-file `openhealth` skill. The Go client package remains available for
developers who want to embed the local runtime directly.

Normal use does not require a daemon, bound port, hosted service, or Go
toolchain on the client machine.

## Install the agent app

Download the `openhealth` archive for your platform from a tagged
GitHub Release, unpack it, and put the binary on `PATH`.

Install the skill by placing the released `SKILL.md` at:

```text
.agents/skills/openhealth/SKILL.md
```

The skill calls these production runner domains:

```bash
openhealth weight
openhealth blood-pressure
openhealth medications
openhealth labs
```

The runner reads JSON from stdin and writes JSON to stdout. The skill is the
client-agent contract for validation, request shapes, and answer hygiene.

## Install in your Go project

Developers can also import the local runtime directly:

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

	"github.com/yazanabuashour/openhealth/client"
)

func main() {
	api, err := client.OpenLocal(client.LocalConfig{})
	if err != nil {
		log.Fatal(err)
	}
	defer api.Close()

	summary, err := api.GetHealthSummaryWithResponse(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if summary.JSON200 == nil {
		log.Fatalf("unexpected status: %s", summary.Status())
	}

	fmt.Printf("active medications=%d\n", summary.JSON200.ActiveMedicationCount)
}
```

`client.OpenLocal(...)` opens SQLite locally, runs migrations, and routes requests to the generated handler in-process. Use `client.NewDefault(baseURL)` only when you intentionally want to talk to an explicit HTTP server.

For common local weight tasks, prefer the ergonomic helper methods on the local
client over generated OpenAPI method names:

```go
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

fmt.Printf("%s %.1f lb %s\n", result.Entry.RecordedAt.Format(time.DateOnly), result.Entry.Value, result.Status)
```

## Local storage

By default, the local runtime stores its SQLite database under `${XDG_DATA_HOME:-~/.local/share}/openhealth/openhealth.db`.

Override the default location with either:

- `client.LocalConfig{DataDir: "..."}`
- `client.LocalConfig{DatabasePath: "..."}`
- `OPENHEALTH_DATA_DIR`
- `OPENHEALTH_DATABASE_PATH`

The database path override wins over the data directory override.

## Contributing and maintainer setup

Repository development still uses the full local toolchain:

```bash
mise install
OPENHEALTH_DATA_DIR="$(mktemp -d)" mise exec -- go run ./examples/client_summary
mise exec -- go generate ./...
mise exec -- gofmt -w .
mise exec -- golangci-lint run
mise exec -- go test ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for contributor expectations and [docs/maintainers.md](docs/maintainers.md) for maintainer-only workflow details.

## Repository contents

- [openapi/openapi.yaml](openapi/openapi.yaml) defines the API contract that generates the server and client bindings.
- [client](client) contains the checked-in generated Go client plus local runtime bootstrap helpers.
- [cmd/openhealth](cmd/openhealth) contains the production JSON runner binary used by the skill.
- [examples/client_summary](examples/client_summary) shows a minimal no-daemon consumer program that imports the generated client.
- [skills/openhealth/SKILL.md](skills/openhealth/SKILL.md) is the complete shipped OpenHealth skill payload.
- [CONTRIBUTING.md](CONTRIBUTING.md) explains how outside contributors should propose changes.
- [SECURITY.md](SECURITY.md) explains how to report vulnerabilities privately and what response timing to expect.
- [docs/maintainers.md](docs/maintainers.md) documents Beads-based maintainer workflow and repo administration notes.
- [docs/release-verification.md](docs/release-verification.md) explains how to verify published source releases.
- [docs/agent-evals.md](docs/agent-evals.md) explains how to evaluate production agent workflows without mixing comparison variants into the production skill.
- [LICENSE](LICENSE) defines the project license.

## Release contract

The production release deliverables are:

- platform archives for the `openhealth` binary
- the single-file `openhealth` skill archive
- the Go module import path rooted at `github.com/yazanabuashour/openhealth`
- the generated client package at `github.com/yazanabuashour/openhealth/client`
- the local in-process runtime surfaced through `client.OpenLocal(...)`

The release workflow is built around semantic version tags in the `v0.y.z`
range. Each tagged GitHub Release publishes binary archives, the skill archive,
a canonical source archive, SHA256 checksums, an SPDX SBOM, and GitHub
attestations for release verification. The repository does not publish a hosted
service deployment target.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution expectations and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community standards.
