# Contributing

Outside contributors do not need Beads to contribute to this repository.

## Project Shape

This repository exposes a production `openhealth` runner binary in
`cmd/openhealth`, a single-file OpenHealth skill in `skills/openhealth/SKILL.md`,
and a direct-local Go package in `client`.

Changes to the Go surface must keep runtime behavior, setup docs, and CI checks
aligned.

## Local Setup

Maintainers prefer:

```bash
mise install
```

Outside contributors may use their own tooling if they can satisfy the
repository checks. Beads and Dolt are maintainer-only tools and are not required
to open, review, or merge pull requests.

Contributors should be able to run:

```bash
printf '%s\n' '{"action":"list_weights","list_mode":"latest"}' | \
  OPENHEALTH_DATABASE_PATH="$(mktemp -d)/openhealth.db" mise exec -- go run ./cmd/openhealth weight
mise exec -- go test ./cmd/openhealth
./scripts/validate-agent-skill.sh skills/openhealth
mise exec -- gofmt -w .
mise exec -- golangci-lint run
mise exec -- go test ./...
```

`golangci-lint` is pinned by `mise.toml`; run it through `mise exec` instead
of relying on a global binary.

If a change touches `skills/openhealth/SKILL.md`, run
`./scripts/validate-agent-skill.sh skills/openhealth` before opening the pull
request.

## Pull Request Expectations

- Keep changes reviewable without access to Beads state.
- Update repository docs when the public contract changes.
- Do not commit credentials, private infrastructure details, or sensitive sample data.
- Route security issues through the private process in [SECURITY.md](SECURITY.md), not through public issues or pull requests.

## Checks and Review Rules

Pull request checks validate repository policy, Agent Skill metadata shape, Go
formatting, Go linting, unit tests, and dependency-review safety.

Pull requests that touch Go code are expected to leave the repository in a
runnable, formatted, lint-clean, and test-clean state.

## Support and Compatibility

Before `1.0`, compatibility is best effort and may change between releases. The
production install story is the `openhealth` runner plus the single-file
OpenHealth skill. The Go package is a developer/source import path for direct
local embedding.

Go `1.26.2` is required for repository development and CI validation on
`ubuntu-latest`. Routine client-agent use should not require a Go toolchain.
OpenHealth does not promise a hosted deployment target or remote HTTP API
contract.

Maintainer workflow notes live in [docs/maintainers.md](docs/maintainers.md).
