# oh-5yr Agent Eval Results

Date: maturity-throughput-speed-shared

Harness: `codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `4`

Cache mode: `shared`

Cache prewarm seconds: `20.18`

Harness elapsed seconds: `47.42`

Effective parallel speedup: `3.64x`

Parallel efficiency: `0.91`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-maturity-throughput-speed-shared.json`

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
| `production` | `add-two` | pass | pass | pass | 1 | 2 | 13.11 | in 45526 / cached 41728 / out 705 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 1 | 2 | 7.11 | in 45075 / cached 41216 / out 333 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 2 | 3 | 13.41 | in 68614 / cached 64128 / out 815 | no | no | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 1 | 2 | 6.83 | in 45066 / cached 41216 / out 290 | no | no | no | no | no | no |
| `production` | `mixed-add-latest` | pass | pass | pass | 4 | 4 | 22.10 | in 115529 / cached 109952 / out 1107 | no | no | no | no | no | no |
| `production` | `mixed-bounded-range` | pass | pass | pass | 2 | 3 | 13.98 | in 91431 / cached 86528 / out 710 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 2 | 4 | 14.31 | in 91459 / cached 86528 / out 638 | no | no | no | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 2 | 3 | 19.04 | in 67890 / cached 63616 / out 438 | no | no | no | no | yes | no |
| `cli` | `bp-add-two` | pass | pass | pass | 4 | 3 | 19.01 | in 138877 / cached 129280 / out 866 | no | no | no | no | yes | no |
| `cli` | `bp-latest-only` | pass | pass | pass | 2 | 4 | 13.97 | in 102585 / cached 93696 / out 608 | no | yes | yes | no | yes | no |
| `cli` | `mixed-add-latest` | pass | pass | pass | 5 | 4 | 18.39 | in 142280 / cached 135936 / out 1102 | no | no | no | no | yes | no |
| `cli` | `mixed-bounded-range` | pass | pass | pass | 3 | 3 | 11.50 | in 68305 / cached 63616 / out 688 | no | no | no | no | yes | no |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.27 |
| install_variant | 0.00 |
| warm_cache | 0.00 |
| seed_db | 0.05 |
| agent_run | 172.76 |
| parse_metrics | 0.00 |
| verify | 0.01 |
| total_job_time | 173.19 |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-oh-967-smoke.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | -3 | -3 | -20.10 | -2845 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | -1 | -2 | -20.89 | -1864 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | same_pass | -1 | -2 | -20.13 | -2040 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-latest-only` | same_pass | -2 | -1 | -12.90 | -1431 | same_no | same_no | same_no | same_no | same_no | same_no |
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
| `total_tools_less_than_or_equal_cli` | pass | production tools 11 vs cli tools 18 |
| `minimum_scenarios_at_or_below_cli` | pass | 6 scenarios at or below CLI tools; required 5 of 6 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |
| `non_cached_token_majority` | pass | 5 scenarios with lower non-cached input tokens; required 4 of 6; missing usage: none |
| `non_cached_token_total_less_than_or_equal_cli` | pass | production non-cached input tokens 26473 vs cli 38724; missing usage: none |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 1 | 2 | -1 |
| `latest-only` | 1 | 2 | -1 |
| `bp-add-two` | 2 | 4 | -2 |
| `bp-latest-only` | 1 | 2 | -1 |
| `mixed-add-latest` | 4 | 5 | -1 |
| `mixed-bounded-range` | 2 | 3 | -1 |

## Metric Evidence

- `cli/add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb && go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb && go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/bp-add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 118 --diastolic 76'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-latest-only` broad repo search: `/bin/zsh -lc "sed -n '1,220p' .agents/skills/openhealth/SKILL.md && printf '\\n---\\n' && rg -n \"blood-pressure|weight|data path|local data|OPENHEALTH\" -S . -g '"'!**/node_modules/**'"'"`.
- `cli/bp-latest-only` generated path from broad search: `/bin/zsh -lc "sed -n '1,220p' .agents/skills/openhealth/SKILL.md && printf '\\n---\\n' && rg -n \"blood-pressure|weight|data path|local data|OPENHEALTH\" -S . -g '"'!**/node_modules/**'"'"`.
- `cli/bp-latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/mixed-add-latest` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-31 --value 150.8 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-31 --systolic 119 --diastolic 77 --pulse 62'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/mixed-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.

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
