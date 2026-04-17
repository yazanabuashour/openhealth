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
- Do not normalize year-first slash dates, such as `2026/03/31`; reject them
  instead of rewriting them. Explicit month/day/year dates with a year, such as
  `03/29/2026`, may be converted to `YYYY-MM-DD`.
- Do not convert unsupported units. The CLI accepts `lb` only.
- Do not write non-positive, missing, or otherwise invalid values.
- Treat direct SQLite access as a read-only fallback. Do not manually insert,
  update, delete, or soft-delete rows in SQLite for write or correction tasks.
  If the CLI cannot run because Go modules, runtime setup, or network access is
  unavailable, report that blocker instead of mutating the database directly.

## Add Weight Entries

Use `weight add` with an ISO date, positive value, and `lb` unit. The command
is idempotent for a date: if a same-day weight already exists, use `weight add`
with the corrected value and the command will update the existing row safely.

```bash
go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb
go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb
```

The command uses the configured local OpenHealth data path. It prints the stored
date, value, unit, and write status.

## Read Weight Entries

Use `weight list` for history. Results are newest first. For bounded date-range
requests, include `--from` and `--to`, then report every row printed by that
bounded command. Do not report only the latest row and do not append rows from
outside the requested range. Do not mention excluded dates at all, even to say
they were excluded.

```bash
go run ./cmd/openhealth weight list --limit 25
go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30
```

If the bounded command prints `2026-03-30` and `2026-03-29`, report both rows,
newest first.

If the CLI cannot run and you inspect SQLite directly as a fallback, keep the
same date-only bounds instead of listing all history:

```sql
SELECT substr(recorded_at, 1, 10) AS date, value, unit
FROM health_weight_entry
WHERE deleted_at IS NULL
  AND substr(recorded_at, 1, 10) >= '2026-03-29'
  AND substr(recorded_at, 1, 10) <= '2026-03-30'
ORDER BY recorded_at DESC, id DESC;
```
