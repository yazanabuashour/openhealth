# openhealth

## Quick start

The primary user-facing runtime is:

- the generated Go client in `client`
- a local in-process runtime opened with `client.OpenLocal(...)`
- the contract source of truth in `openapi/openapi.yaml`

```bash
mise install
OPENHEALTH_DATA_DIR="$(mktemp -d)" mise exec -- go run ./examples/client_summary
```

Validate local changes with:

```bash
mise exec -- go generate ./...
mise exec -- gofmt -w .
mise exec -- golangci-lint run
mise exec -- go test ./...
```

OpenHealth keeps the OpenAPI-generated client surface without requiring a host-level daemon, bound port, or background service. The client bootstrap opens SQLite locally, runs migrations, and routes requests to the generated handler in-process.

Minimal usage from Go:

```go
api, err := client.OpenLocal(client.LocalConfig{})
if err != nil {
  return err
}
defer api.Close()

summary, err := api.GetHealthSummaryWithResponse(ctx)
```

## Local storage

By default, the local runtime stores its SQLite database under `${XDG_DATA_HOME:-~/.local/share}/openhealth/openhealth.db`.

Override the default location with either:

- `client.LocalConfig{DataDir: "..."}`
- `client.LocalConfig{DatabasePath: "..."}`
- `OPENHEALTH_DATA_DIR`
- `OPENHEALTH_DATABASE_PATH`

The database path override wins over the data directory override.

## Maintainer CLI

The CLI remains available for maintainers who want an explicit HTTP server for debugging or contract inspection:

```bash
mise exec -- go run ./cmd/openhealth migrate
mise exec -- go run ./cmd/openhealth serve
```

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

The current source-level deliverables are:

- the Go module import path rooted at `github.com/yazanabuashour/openhealth`
- the generated client package at `github.com/yazanabuashour/openhealth/client`
- the local in-process runtime surfaced through `client.OpenLocal(...)`

The maintainer CLI under `cmd/openhealth` remains part of the repository for debugging and contract inspection, but it is not the primary end-user install surface.

Releases remain GitHub Releases with semantic version tags in the `v0.y.z` range. Each tagged release publishes a canonical source archive, SHA256 checksums, an SPDX SBOM, and GitHub attestations for release verification. The repository does not currently publish downloadable platform binaries or a hosted service deployment target.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution expectations and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community standards.
