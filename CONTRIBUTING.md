# Contributing

Outside contributors do not need Beads to contribute to this repository.

## Current project shape

This repository now includes a small runnable Go CLI at `cmd/openhealth`. It is still pre-`1.0`, does not publish packages or deployment artifacts, and keeps its public contract intentionally small.

Changes to the Go surface must keep the runtime, setup instructions, and validation commands truthful in both docs and CI.

## Local setup

Maintainers prefer:

```bash
mise install
```

Outside contributors may use their own local tooling if they can satisfy the repository checks.

Beads and Dolt are maintainer-only tools. They are optional for outside contributors and are not required to open, review, or merge pull requests.

For the current Go surface, contributors should be able to run:

```bash
mise exec -- go run ./cmd/openhealth
mise exec -- gofmt -w .
mise exec -- golangci-lint run
mise exec -- go test ./...
```

## Pull request expectations

- Keep changes reviewable without access to Beads state.
- Update repository docs when the public contract changes.
- Do not commit credentials, private infrastructure details, or sensitive sample data.
- Route security issues through the private process in [SECURITY.md](SECURITY.md), not through public issues or pull requests.

## Checks and review rules

Current pull request checks validate repository policy, Go formatting, Go linting, unit tests, and dependency-review safety.

Pull requests that touch Go code are expected to leave the repository in a runnable, formatted, lint-clean, and test-clean state.

## Support and compatibility

Before `0.1.0`, compatibility is best effort and may change between releases. The current supported runtime surface is Go `1.26.2`, with CI validating the repository on `ubuntu-latest`. No deployment target, hosted service, or packaged binary distribution is promised yet.

Maintainer workflow notes live in [docs/maintainers.md](docs/maintainers.md).
