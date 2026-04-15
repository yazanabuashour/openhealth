# openhealth

OpenHealth is a Go SDK for a local-first health runtime. The default install surface is the generated client package in `client`, opened in-process with `client.OpenLocal(...)`. Normal use does not require a daemon, bound port, or background service.

## Install in your Go project

Install the first tagged release of the module, then import the client package from it:

```bash
go get github.com/yazanabuashour/openhealth@v0.1.0
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

## Local storage

By default, the local runtime stores its SQLite database under `${XDG_DATA_HOME:-~/.local/share}/openhealth/openhealth.db`.

Override the default location with either:

- `client.LocalConfig{DataDir: "..."}`
- `client.LocalConfig{DatabasePath: "..."}`
- `OPENHEALTH_DATA_DIR`
- `OPENHEALTH_DATABASE_PATH`

The database path override wins over the data directory override.

## Debugging with a standalone server

The standalone CLI remains available for maintainer debugging or contract inspection, but it is not the primary end-user install surface:

```bash
mise exec -- go run ./cmd/openhealth serve
```

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

Maintainers who need explicit database or HTTP debugging can also run:

```bash
mise exec -- go run ./cmd/openhealth migrate
mise exec -- go run ./cmd/openhealth serve
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for contributor expectations and [docs/maintainers.md](docs/maintainers.md) for maintainer-only workflow details.

## Repository contents

- [openapi/openapi.yaml](openapi/openapi.yaml) defines the API contract that generates the server and client bindings.
- [client](client) contains the checked-in generated Go client plus local runtime bootstrap helpers.
- [cmd/openhealth](cmd/openhealth) contains the maintainer/debug CLI with explicit `migrate` and `serve` commands.
- [examples/client_summary](examples/client_summary) shows a minimal no-daemon consumer program that imports the generated client.
- [skills/openhealth/SKILL.md](skills/openhealth/SKILL.md) provides optional agent-oriented install and usage guidance on top of the OpenAPI contract.
- [CONTRIBUTING.md](CONTRIBUTING.md) explains how outside contributors should propose changes.
- [SECURITY.md](SECURITY.md) explains how to report vulnerabilities privately and what response timing to expect.
- [docs/maintainers.md](docs/maintainers.md) documents Beads-based maintainer workflow and repo administration notes.
- [docs/release-verification.md](docs/release-verification.md) explains how to verify published source releases.
- [LICENSE](LICENSE) defines the project license.

## Release contract

The source-level deliverables are:

- the Go module import path rooted at `github.com/yazanabuashour/openhealth`
- the generated client package at `github.com/yazanabuashour/openhealth/client`
- the local in-process runtime surfaced through `client.OpenLocal(...)`

The maintainer CLI under `cmd/openhealth` remains part of the repository for debugging and contract inspection, but it is not the primary end-user install surface.

The release workflow is built around semantic version tags in the `v0.y.z` range. Each tagged GitHub Release publishes a canonical source archive, SHA256 checksums, an SPDX SBOM, and GitHub attestations for release verification. The repository does not publish downloadable platform binaries or a hosted service deployment target.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution expectations and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community standards.
