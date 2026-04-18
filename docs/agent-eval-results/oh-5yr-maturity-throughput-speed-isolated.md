# oh-5yr Agent Eval Results

Date: maturity-throughput-speed-isolated

Harness: `codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `4`

Cache mode: `isolated`

Harness elapsed seconds: `159.11`

Effective parallel speedup: `3.18x`

Parallel efficiency: `0.80`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-maturity-throughput-speed-isolated.json`

Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.

## History Isolation

- Status: `passed`
- Single-turn ephemeral runs: `12`.
- Multi-turn persisted sessions: `0` sessions / `0` turns.
- The Codex desktop app was not used for eval prompts.
- New Codex session files referencing `<run-root>`: `0`.

## Results

| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | pass | pass | pass | 1 | 2 | 38.59 | in 45168 / cached 41216 / out 395 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 1 | 2 | 34.26 | in 44952 / cached 41216 / out 287 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 1 | 4 | 38.03 | in 161604 / cached 156800 / out 963 | no | no | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 2 | 5 | 42.44 | in 160473 / cached 154752 / out 929 | no | no | no | no | no | no |
| `production` | `mixed-add-latest` | pass | pass | pass | 4 | 7 | 45.42 | in 209398 / cached 203136 / out 1260 | no | no | no | no | no | no |
| `production` | `mixed-bounded-range` | pass | pass | pass | 2 | 3 | 36.16 | in 137623 / cached 131840 / out 817 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 4 | 3 | 35.74 | in 186222 / cached 179712 / out 1061 | no | no | no | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 2 | 4 | 32.62 | in 90931 / cached 86016 / out 507 | no | no | no | no | yes | no |
| `cli` | `bp-add-two` | pass | pass | pass | 4 | 5 | 51.92 | in 209951 / cached 198528 / out 1085 | no | no | no | no | yes | no |
| `cli` | `bp-latest-only` | pass | pass | pass | 4 | 9 | 49.18 | in 209850 / cached 203648 / out 1261 | no | no | no | no | yes | no |
| `cli` | `mixed-add-latest` | pass | pass | pass | 7 | 9 | 54.15 | in 214786 / cached 201600 / out 1901 | no | no | no | no | yes | no |
| `cli` | `mixed-bounded-range` | pass | pass | pass | 3 | 6 | 47.72 | in 139961 / cached 133888 / out 1194 | no | no | no | no | yes | no |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.33 |
| install_variant | 0.00 |
| warm_cache | 113.86 |
| seed_db | 0.05 |
| agent_run | 506.23 |
| parse_metrics | 0.00 |
| verify | 0.00 |
| total_job_time | 620.55 |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-oh-967-smoke.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | -3 | -3 | +5.38 | -2691 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | -1 | -2 | +6.26 | -1987 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | same_pass | -2 | -1 | +4.49 | -1722 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-latest-only` | same_pass | -1 | +2 | +22.71 | +440 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `mixed-add-latest` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mixed-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `add-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `latest-only` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-add-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-latest-only` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mixed-add-latest` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mixed-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |

## Production Stop-Loss

- Triggered: `no`
- Recommendation: `continue_production_hardening`

## Code-First CLI Comparison

- Candidate: `production`
- Baseline: `cli`
- Beats CLI: `yes`
- Recommendation: `prefer_agentops_production_for_routine_openhealth_operations`

| Criterion | Result | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | pass | 6/6 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `validation_scenarios_are_final_answer_only` | pass | validation scenarios used no tools, no command executions, and at most one assistant answer |
| `total_tools_less_than_or_equal_cli` | pass | production tools 11 vs cli tools 24 |
| `minimum_scenarios_at_or_below_cli` | pass | 6 scenarios at or below CLI tools; required 5 of 6 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |
| `non_cached_token_majority` | pass | 6 scenarios with lower non-cached input tokens; required 4 of 6; missing usage: none |
| `non_cached_token_total_less_than_or_equal_cli` | pass | production non-cached input tokens 30258 vs cli 48309; missing usage: none |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 1 | 4 | -3 |
| `latest-only` | 1 | 2 | -1 |
| `bp-add-two` | 1 | 4 | -3 |
| `bp-latest-only` | 2 | 4 | -2 |
| `mixed-add-latest` | 4 | 7 | -3 |
| `mixed-bounded-range` | 2 | 3 | -1 |

## Metric Evidence

- `cli/add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/bp-add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 118 --diastolic 76'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 25'`.
- `cli/bp-latest-only` openhealth CLI: `/bin/zsh -lc 'timeout 20s go run ./cmd/openhealth blood-pressure list --limit 1; status=$?; echo EXIT:$status'`; `/bin/zsh -lc "perl -e 'alarm 20; exec @ARGV' go run ./cmd/openhealth blood-pressure list --limit 1"`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/mixed-add-latest` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-31 --value 150.8 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-31 --systolic 119 --diastolic 77 --pulse 62'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/mixed-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.

## Scenario Notes

- `production/add-two`: turn 1: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/add-two/turn-1/events.jsonl`.
- `production/latest-only`: turn 1: expected unchanged seed rows and assistant output limited to latest row 2026-03-30; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/latest-only/turn-1/events.jsonl`.
- `production/bp-add-two`: turn 1: expected exactly two newest-first blood-pressure rows; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/bp-add-two/turn-1/events.jsonl`.
- `production/bp-latest-only`: turn 1: expected unchanged seed rows and assistant output limited to latest blood-pressure row 2026-03-30; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/bp-latest-only/turn-1/events.jsonl`.
- `production/mixed-add-latest`: turn 1: expected latest weight and blood-pressure rows for 2026-03-31; observed weights [2026-03-31 150.8 lb] and blood pressures [2026-03-31 119/77 pulse 62] Raw event reference: `<run-root>/production/mixed-add-latest/turn-1/events.jsonl`.
- `production/mixed-bounded-range`: turn 1: expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/mixed-bounded-range/turn-1/events.jsonl`.
- `cli/add-two`: turn 1: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/cli/add-two/turn-1/events.jsonl`.
- `cli/latest-only`: turn 1: expected unchanged seed rows and assistant output limited to latest row 2026-03-30; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/cli/latest-only/turn-1/events.jsonl`.
- `cli/bp-add-two`: turn 1: expected exactly two newest-first blood-pressure rows; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/bp-add-two/turn-1/events.jsonl`.
- `cli/bp-latest-only`: turn 1: expected unchanged seed rows and assistant output limited to latest blood-pressure row 2026-03-30; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/bp-latest-only/turn-1/events.jsonl`.
- `cli/mixed-add-latest`: turn 1: expected latest weight and blood-pressure rows for 2026-03-31; observed weights [2026-03-31 150.8 lb] and blood pressures [2026-03-31 119/77 pulse 62] Raw event reference: `<run-root>/cli/mixed-add-latest/turn-1/events.jsonl`.
- `cli/mixed-bounded-range`: turn 1: expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/mixed-bounded-range/turn-1/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands with the configured Go cache mode`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
