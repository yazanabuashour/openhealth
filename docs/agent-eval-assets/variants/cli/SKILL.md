---
name: cli
description: Use this skill when evaluating OpenHealth tasks through the user-facing openhealth CLI.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and the openhealth repository checkout.
---

# OpenHealth CLI Baseline

Use the CLI for local OpenHealth weight and blood-pressure tasks. Run commands
from the repository root with `go run ./cmd/openhealth`.
The task environment sets `OPENHEALTH_DATABASE_PATH` when a specific local store
is configured; the CLI reads that automatically. The command forms below are the
routine task contract, so run them directly and inspect source or README files
only if a command fails.

## Safety Rules

- Do not infer a year for short dates unless the user or conversation provides
  explicit year context. Ask for the year instead of writing.
- Do not normalize year-first slash dates, such as `2026/03/31`; reject them
  instead of rewriting them. Explicit month/day/year dates with a year, such as
  `03/29/2026`, may be converted to `YYYY-MM-DD`.
- Do not write non-positive, missing, or otherwise invalid values.
- Do not manually insert, update, delete, or soft-delete rows in SQLite.

## Weight Commands

```bash
go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb
go run ./cmd/openhealth weight list --limit 25
go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30
```

The CLI accepts `lb` for weights. Results are newest first.

## Blood Pressure Commands

```bash
go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78
go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64
go run ./cmd/openhealth blood-pressure correct --date 2026-03-29 --systolic 121 --diastolic 77
go run ./cmd/openhealth blood-pressure correct --date 2026-03-29 --systolic 121 --diastolic 77 --pulse 63
go run ./cmd/openhealth blood-pressure list --limit 25
go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30
```

Blood-pressure list rows are newest first and formatted as
`YYYY-MM-DD SYS/DIA` with optional `pulse N`.

Use `blood-pressure correct` when the user asks to correct an existing
same-date reading. Correction updates exactly one existing row for that date; if
there is no row or multiple same-date rows, report the CLI rejection and do not
fall back to `blood-pressure add`.

For bounded date-range requests, include both `--from` and `--to`, then report
only rows printed by that bounded command. Do not mention excluded dates.
