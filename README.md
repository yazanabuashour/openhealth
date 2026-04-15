# openhealth

## Quick start

The repository now includes a small source-only Go CLI at `cmd/openhealth`.

```bash
mise install
mise exec -- go run ./cmd/openhealth
```

Validate local changes with:

```bash
mise exec -- gofmt -w .
mise exec -- golangci-lint run
mise exec -- go test ./...
```

## Repository contents

- [cmd/openhealth](cmd/openhealth) contains the first runnable Go entrypoint for the repository.
- [CONTRIBUTING.md](CONTRIBUTING.md) explains how outside contributors should propose changes.
- [SECURITY.md](SECURITY.md) explains how to report vulnerabilities privately and what response timing to expect.
- [docs/maintainers.md](docs/maintainers.md) documents Beads-based maintainer workflow and repo administration notes.
- [LICENSE](LICENSE) defines the project license.

## Release contract

The initial runnable surface is the source tree under `cmd/openhealth`. Releases remain GitHub Releases with semantic version tags in the `0.y.z` range. Release notes are generated from protected tags, and the repository does not currently publish packages or downloadable build artifacts.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution expectations and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community standards.
