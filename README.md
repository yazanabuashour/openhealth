# openhealth

## Quick start

The repository exposes:

- an OpenAPI-first health service at `cmd/openhealth`
- a generated Go client in `client`
- the contract source of truth in `openapi/openapi.yaml`

```bash
mise install
mise exec -- go run ./cmd/openhealth migrate
mise exec -- go run ./cmd/openhealth serve
```

Validate local changes with:

```bash
mise exec -- go generate ./...
mise exec -- gofmt -w .
mise exec -- golangci-lint run
mise exec -- go test ./...
```

Try the generated client against a local server with:

```bash
mise exec -- go run ./examples/client_summary
```

## Repository contents

- [openapi/openapi.yaml](openapi/openapi.yaml) defines the API contract that generates the server and client bindings.
- [cmd/openhealth](cmd/openhealth) contains the runnable CLI with explicit `migrate` and `serve` commands.
- [client](client) contains the checked-in generated Go client plus a small helper for timeout and safe read retries.
- [examples/client_summary](examples/client_summary) shows a minimal consumer program that imports the generated client.
- [skills/openhealth/SKILL.md](skills/openhealth/SKILL.md) provides an optional OpenClaw-oriented onboarding artifact for agents.
- [CONTRIBUTING.md](CONTRIBUTING.md) explains how outside contributors should propose changes.
- [SECURITY.md](SECURITY.md) explains how to report vulnerabilities privately and what response timing to expect.
- [docs/maintainers.md](docs/maintainers.md) documents Beads-based maintainer workflow and repo administration notes.
- [LICENSE](LICENSE) defines the project license.

## Release contract

The current source-level deliverables are:

- the Go module import path rooted at `github.com/yazanabuashour/openhealth`
- the generated client package at `github.com/yazanabuashour/openhealth/client`
- the service entrypoint under `cmd/openhealth`

Releases remain GitHub Releases with semantic version tags in the `0.y.z` range. Release notes are generated from protected tags, and the repository does not currently publish downloadable binaries or separate package artifacts.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution expectations and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community standards.
