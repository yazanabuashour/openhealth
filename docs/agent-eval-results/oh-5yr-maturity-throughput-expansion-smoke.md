# oh-5yr Agent Eval Results

Date: maturity-throughput-expansion-smoke

Harness: `codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `4`

Cache mode: `shared`

Cache prewarm seconds: `18.40`

Harness elapsed seconds: `66.94`

Effective parallel speedup: `2.38x`

Parallel efficiency: `0.60`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-maturity-throughput-expansion-smoke.json`

Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.

## History Isolation

- Status: `passed`
- Single-turn ephemeral runs: `4`.
- Multi-turn persisted sessions: `4` sessions / `8` turns.
- The Codex desktop app was not used for eval prompts.
- New Codex session files referencing `<run-root>`: `4`.

## Results

| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |
| `production` | `mixed-add-latest` | pass | pass | pass | 5 | 4 | 19.63 | in 142064 / cached 135936 / out 1005 | no | no | no | no | no | no |
| `production` | `mixed-bounded-range` | pass | pass | pass | 3 | 3 | 13.51 | in 95266 / cached 83968 / out 702 | no | no | no | no | no | no |
| `production` | `mt-weight-clarify-then-add` | pass | pass | pass | 1 | 3 | 12.90 | in 90539 / cached 82944 / out 822 | no | no | no | no | no | no |
| `production` | `mt-mixed-latest-then-correct` | pass | pass | pass | 4 | 6 | 25.80 | in 253485 / cached 242816 / out 1712 | no | no | no | no | no | no |
| `cli` | `mixed-add-latest` | pass | pass | pass | 5 | 4 | 13.74 | in 92176 / cached 87040 / out 1070 | no | no | no | no | yes | no |
| `cli` | `mixed-bounded-range` | pass | pass | pass | 3 | 3 | 11.22 | in 68280 / cached 61056 / out 656 | no | no | no | no | yes | no |
| `cli` | `mt-weight-clarify-then-add` | pass | pass | pass | 3 | 5 | 20.39 | in 136594 / cached 128256 / out 1053 | no | no | no | no | yes | no |
| `cli` | `mt-mixed-latest-then-correct` | pass | pass | pass | 15 | 8 | 42.11 | in 359732 / cached 312704 / out 3632 | no | yes | yes | no | yes | no |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.22 |
| install_variant | 0.00 |
| warm_cache | 0.00 |
| seed_db | 0.04 |
| agent_run | 159.30 |
| parse_metrics | 0.00 |
| verify | 0.00 |
| total_job_time | 159.64 |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-maturity-throughput-expansion-smoke.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `mixed-add-latest` | same_pass | +1 | +0 | +1.29 | +799 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `mixed-bounded-range` | same_pass | +1 | +0 | +0.28 | +7064 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `mt-weight-clarify-then-add` | same_pass | +0 | -2 | -13.38 | -2085 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `mt-mixed-latest-then-correct` | same_pass | +0 | +1 | +1.73 | +142 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `mixed-add-latest` | same_pass | +0 | +0 | -2.96 | -1 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `mixed-bounded-range` | same_pass | +0 | -1 | -4.42 | +2336 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `mt-weight-clarify-then-add` | same_pass | +1 | +0 | -0.51 | -714 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `mt-mixed-latest-then-correct` | fixed | +5 | +0 | +4.78 | +35717 | same_no | regressed_to_yes | regressed_to_yes | same_no | same_yes | same_no |

## Production Stop-Loss

- Triggered: `no`
- Recommendation: `continue_production_hardening`

## Code-First CLI Comparison

- Candidate: `production`
- Baseline: `cli`
- Beats CLI: `no`
- Recommendation: `continue_cli_for_routine_openhealth_operations`

| Criterion | Result | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | pass | 4/4 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `validation_scenarios_are_final_answer_only` | pass | validation scenarios used no tools, no command executions, and at most one assistant answer |
| `total_tools_less_than_or_equal_cli` | pass | production tools 13 vs cli tools 26 |
| `minimum_scenarios_at_or_below_cli` | pass | 4 scenarios at or below CLI tools; required 4 of 4 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |
| `non_cached_token_majority` | fail | 2 scenarios with lower non-cached input tokens; required 3 of 4; missing usage: none |
| `non_cached_token_total_less_than_or_equal_cli` | pass | production non-cached input tokens 35690 vs cli 67726; missing usage: none |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `mixed-add-latest` | 5 | 5 | +0 |
| `mixed-bounded-range` | 3 | 3 | +0 |
| `mt-weight-clarify-then-add` | 1 | 3 | -2 |
| `mt-mixed-latest-then-correct` | 4 | 15 | -11 |

## Metric Evidence

- `cli/mixed-add-latest` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-31 --systolic 119 --diastolic 77 --pulse 62'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-31 --value 150.8 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 25'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/mixed-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/mt-weight-clarify-then-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/mt-mixed-latest-then-correct` broad repo search: `/bin/zsh -lc 'rg -n "same date|duplicate|upsert|weight add|blood-pressure add|INSERT|ON CONFLICT|latest" cmd internal .'`.
- `cli/mt-mixed-latest-then-correct` generated path from broad search: `/bin/zsh -lc 'rg -n "same date|duplicate|upsert|weight add|blood-pressure add|INSERT|ON CONFLICT|latest" cmd internal .'`.
- `cli/mt-mixed-latest-then-correct` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth --help'`.

## Turn Details

- `production/mt-weight-clarify-then-add` turn 1: exit `0`, tools `0`, assistant calls `1`, wall `4.68`, raw `<run-root>/production/mt-weight-clarify-then-add/turn-1/events.jsonl`.
- `production/mt-weight-clarify-then-add` turn 2: exit `0`, tools `1`, assistant calls `2`, wall `8.22`, raw `<run-root>/production/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `production/mt-mixed-latest-then-correct` turn 1: exit `0`, tools `2`, assistant calls `4`, wall `13.66`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-1/events.jsonl`.
- `production/mt-mixed-latest-then-correct` turn 2: exit `0`, tools `2`, assistant calls `2`, wall `12.14`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
- `cli/mt-weight-clarify-then-add` turn 1: exit `0`, tools `0`, assistant calls `1`, wall `5.25`, raw `<run-root>/cli/mt-weight-clarify-then-add/turn-1/events.jsonl`.
- `cli/mt-weight-clarify-then-add` turn 2: exit `0`, tools `3`, assistant calls `4`, wall `15.14`, raw `<run-root>/cli/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `cli/mt-mixed-latest-then-correct` turn 1: exit `0`, tools `3`, assistant calls `3`, wall `12.42`, raw `<run-root>/cli/mt-mixed-latest-then-correct/turn-1/events.jsonl`.
- `cli/mt-mixed-latest-then-correct` turn 2: exit `0`, tools `12`, assistant calls `5`, wall `29.69`, raw `<run-root>/cli/mt-mixed-latest-then-correct/turn-2/events.jsonl`.

## Scenario Notes

- `production/mixed-add-latest`: turn 1: expected latest weight and blood-pressure rows for 2026-03-31; observed weights [2026-03-31 150.8 lb] and blood pressures [2026-03-31 119/77 pulse 62] Raw event reference: `<run-root>/production/mixed-add-latest/turn-1/events.jsonl`.
- `production/mixed-bounded-range`: turn 1: expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/mixed-bounded-range/turn-1/events.jsonl`.
- `production/mt-weight-clarify-then-add`: turn 1: expected no first-turn write and a year clarification; observed weights []; turn 2: expected second-turn write after year clarification; observed weights [2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `production/mt-mixed-latest-then-correct`: turn 1: expected unchanged seed rows and latest mixed-domain answer; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest mixed-domain corrections on 2026-03-30; observed weights [2026-03-30 151.0 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
- `cli/mixed-add-latest`: turn 1: expected latest weight and blood-pressure rows for 2026-03-31; observed weights [2026-03-31 150.8 lb] and blood pressures [2026-03-31 119/77 pulse 62] Raw event reference: `<run-root>/cli/mixed-add-latest/turn-1/events.jsonl`.
- `cli/mixed-bounded-range`: turn 1: expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/mixed-bounded-range/turn-1/events.jsonl`.
- `cli/mt-weight-clarify-then-add`: turn 1: expected no first-turn write and a year clarification; observed weights []; turn 2: expected second-turn write after year clarification; observed weights [2026-03-29 152.2 lb] Raw event reference: `<run-root>/cli/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `cli/mt-mixed-latest-then-correct`: turn 1: expected unchanged seed rows and latest mixed-domain answer; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest mixed-domain corrections on 2026-03-30; observed weights [2026-03-30 151.0 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/mt-mixed-latest-then-correct/turn-2/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands with the configured Go cache mode`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
