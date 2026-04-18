# Contributing

Outside contributors do not need Beads to contribute to this repository.

## Current project shape

This repository exposes a production `openhealth` runner binary in
`cmd/openhealth`, a single-file OpenHealth skill in `skills/openhealth/SKILL.md`,
and an optional direct-local Go SDK in `client`.

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
OPENHEALTH_DATA_DIR="$(mktemp -d)" mise exec -- go run ./examples/client_summary
mise exec -- go test ./cmd/openhealth
./scripts/validate-agent-skill.sh skills/openhealth
mise exec -- gofmt -w .
mise exec -- golangci-lint run
mise exec -- go test ./...
```

If a change touches `skills/openhealth/SKILL.md`, run `./scripts/validate-agent-skill.sh skills/openhealth` before opening the pull request.

## Pull request expectations

- Keep changes reviewable without access to Beads state.
- Update repository docs when the public contract changes.
- Do not commit credentials, private infrastructure details, or sensitive sample data.
- Route security issues through the private process in [SECURITY.md](SECURITY.md), not through public issues or pull requests.

## Checks and review rules

Current pull request checks validate repository policy, the Agent Skill metadata shape, Go formatting, Go linting, unit tests, and dependency-review safety.

Pull requests that touch Go code are expected to leave the repository in a runnable, formatted, lint-clean, and test-clean state.

## Support and compatibility

Before `0.1.0`, compatibility is best effort and may change between releases.
The current production install story is the `openhealth` runner plus the
single-file OpenHealth skill. The optional Go SDK is direct-local only. Go
`1.26.2` is required for repository development and CI validation on
`ubuntu-latest`; routine client-agent use should not require a Go toolchain. No
hosted deployment target, remote HTTP API, or generated API contract is
promised.

Maintainer workflow notes live in [docs/maintainers.md](docs/maintainers.md).
