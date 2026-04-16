---
name: cli
description: Use this skill when evaluating OpenHealth weight tasks through the user-facing openhealth CLI.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and the openhealth repository checkout.
---

# OpenHealth CLI Baseline

Use the CLI for local OpenHealth weight tasks. Run commands from the repository
root with `go run ./cmd/openhealth`.

## Safety Rules

- Do not infer a year for short dates unless the user or conversation provides
  explicit year context. Ask for the year instead of writing.
- Do not convert unsupported units. The CLI accepts `lb` only.
- Do not write non-positive, missing, or otherwise invalid values.

## Add Weight Entries

Use `weight add` with an ISO date, positive value, and `lb` unit.

```bash
go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb
go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb
```

The command uses the configured local OpenHealth data path. It prints the stored
date, value, unit, and write status.

## Read Weight Entries

Use `weight list` for history. Results are newest first.

```bash
go run ./cmd/openhealth weight list --limit 25
go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30
```
