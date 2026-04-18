# oh-5yr Agent Eval Results

Date: oh-23a-smoke

Harness: `codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `4`

Cache mode: `shared`

Cache prewarm seconds: `52.09`

Harness elapsed seconds: `67.18`

Effective parallel speedup: `3.13x`

Parallel efficiency: `0.78`

Reduced JSON artifact: `docs/agent-eval-results/archive/oh-5yr-oh-23a-smoke.json`

Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.

## History Isolation

- Status: `passed`
- Single-turn ephemeral runs: `6`.
- Multi-turn persisted sessions: `4` sessions / `8` turns.
- The Codex desktop app was not used for eval prompts.
- New Codex session files referencing `<run-root>`: `4`.

## Results

| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |
| `production` | `bp-correct-existing` | pass | pass | pass | 2 | 3 | 12.23 | in 68699 / cached 64128 / out 679 | no | no | no | no | no | no |
| `production` | `bp-correct-missing-reject` | pass | pass | pass | 1 | 3 | 13.48 | in 68421 / cached 64128 / out 553 | no | no | no | no | no | no |
| `production` | `bp-correct-ambiguous-reject` | pass | pass | pass | 1 | 3 | 14.74 | in 68760 / cached 64640 / out 759 | no | no | no | no | no | no |
| `production` | `mt-bp-latest-then-correct` | pass | pass | pass | 3 | 5 | 21.34 | in 186468 / cached 176128 / out 1270 | no | no | no | no | no | no |
| `production` | `mt-mixed-latest-then-correct` | pass | pass | pass | 4 | 7 | 31.37 | in 256209 / cached 245888 / out 2334 | no | no | no | no | no | no |
| `cli` | `bp-correct-existing` | pass | pass | pass | 2 | 4 | 16.70 | in 91947 / cached 86528 / out 644 | no | no | no | no | yes | no |
| `cli` | `bp-correct-missing-reject` | pass | pass | pass | 2 | 4 | 13.49 | in 92124 / cached 83968 / out 819 | no | no | no | no | yes | no |
| `cli` | `bp-correct-ambiguous-reject` | pass | pass | pass | 2 | 4 | 15.50 | in 92089 / cached 86528 / out 794 | no | no | no | no | yes | no |
| `cli` | `mt-bp-latest-then-correct` | pass | pass | pass | 4 | 6 | 34.25 | in 207107 / cached 196480 / out 1147 | no | no | no | no | yes | no |
| `cli` | `mt-mixed-latest-then-correct` | pass | pass | pass | 4 | 8 | 36.87 | in 280411 / cached 262144 / out 2095 | no | no | no | no | yes | no |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.38 |
| install_variant | 0.00 |
| warm_cache | 0.00 |
| seed_db | 0.09 |
| agent_run | 209.97 |
| parse_metrics | 0.00 |
| verify | 0.05 |
| total_job_time | 210.53 |

## Comparison

Baseline: `docs/agent-eval-results/archive/oh-5yr-oh-967-smoke.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `bp-correct-existing` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-correct-missing-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-correct-ambiguous-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mt-bp-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mt-mixed-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-correct-existing` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-correct-missing-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-correct-ambiguous-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mt-bp-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mt-mixed-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |

## Metric Notes

- oh-23a intentionally keeps agent-facing readiness scoped to weight and blood pressure; labs and medications remain a separate AgentOps expansion tracked in oh-bng.

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
| `candidate_passes_all_scenarios` | pass | 5/5 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `validation_scenarios_are_final_answer_only` | pass | validation scenarios used no tools, no command executions, and at most one assistant answer |
| `total_tools_less_than_or_equal_cli` | pass | production tools 11 vs cli tools 14 |
| `minimum_scenarios_at_or_below_cli` | pass | 5 scenarios at or below CLI tools; required 4 of 5 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |
| `non_cached_token_majority` | pass | 5 scenarios with lower non-cached input tokens; required 3 of 5; missing usage: none |
| `non_cached_token_total_less_than_or_equal_cli` | pass | production non-cached input tokens 33645 vs cli 48030; missing usage: none |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `bp-correct-existing` | 2 | 2 | +0 |
| `bp-correct-missing-reject` | 1 | 2 | -1 |
| `bp-correct-ambiguous-reject` | 1 | 2 | -1 |
| `mt-mixed-latest-then-correct` | 4 | 4 | +0 |
| `mt-bp-latest-then-correct` | 3 | 4 | -1 |

## Metric Evidence

- `cli/bp-correct-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure correct --date 2026-03-29 --systolic 121 --diastolic 77 --pulse 63 && go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-correct-missing-reject` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-31 --to 2026-03-31'`.
- `cli/bp-correct-ambiguous-reject` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/mt-bp-latest-then-correct` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure correct --date 2026-03-30 --systolic 117 --diastolic 75 --pulse 63'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/mt-mixed-latest-then-correct` openhealth CLI: `/bin/zsh -lc "go run ./cmd/openhealth weight list --limit 1 && printf '\\n' && go run ./cmd/openhealth blood-pressure list --limit 1"`; `/bin/zsh -lc "go run ./cmd/openhealth weight --help && printf '\\n' && go run ./cmd/openhealth blood-pressure --help"`; `/bin/zsh -lc "go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.0 --unit lb && printf '\\n' && go run ./cmd/openhealth blood-pressure correct --date 2026-03-30 --systolic 117 --diastolic 75 --pulse 63 && printf '\\n' && go run ./cmd/openhealth weight list --from 2026-03-30 --to 2026-03-30 && printf '\\n' && go run ./cmd/openhealth blood-pressure list --from 2026-03-30 --to 2026-03-30"`.

## Turn Details

- `production/mt-bp-latest-then-correct` turn 1: exit `0`, tools `2`, assistant calls `3`, wall `12.01`, raw `<run-root>/production/mt-bp-latest-then-correct/turn-1/events.jsonl`.
- `production/mt-bp-latest-then-correct` turn 2: exit `0`, tools `1`, assistant calls `2`, wall `9.33`, raw `<run-root>/production/mt-bp-latest-then-correct/turn-2/events.jsonl`.
- `production/mt-mixed-latest-then-correct` turn 1: exit `0`, tools `2`, assistant calls `4`, wall `17.63`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-1/events.jsonl`.
- `production/mt-mixed-latest-then-correct` turn 2: exit `0`, tools `2`, assistant calls `3`, wall `13.74`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
- `cli/mt-bp-latest-then-correct` turn 1: exit `0`, tools `2`, assistant calls `3`, wall `16.69`, raw `<run-root>/cli/mt-bp-latest-then-correct/turn-1/events.jsonl`.
- `cli/mt-bp-latest-then-correct` turn 2: exit `0`, tools `2`, assistant calls `3`, wall `17.56`, raw `<run-root>/cli/mt-bp-latest-then-correct/turn-2/events.jsonl`.
- `cli/mt-mixed-latest-then-correct` turn 1: exit `0`, tools `2`, assistant calls `4`, wall `15.82`, raw `<run-root>/cli/mt-mixed-latest-then-correct/turn-1/events.jsonl`.
- `cli/mt-mixed-latest-then-correct` turn 2: exit `0`, tools `2`, assistant calls `4`, wall `21.05`, raw `<run-root>/cli/mt-mixed-latest-then-correct/turn-2/events.jsonl`.

## Scenario Notes

- `production/bp-correct-existing`: turn 1: expected corrected 2026-03-29 blood-pressure row with no duplicate; observed [2026-03-29 121/77 pulse 63] Raw event reference: `<run-root>/production/bp-correct-existing/turn-1/events.jsonl`.
- `production/bp-correct-missing-reject`: turn 1: expected unchanged seed row and missing-date correction rejection; observed [2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/bp-correct-missing-reject/turn-1/events.jsonl`.
- `production/bp-correct-ambiguous-reject`: turn 1: expected unchanged duplicate same-date rows and ambiguous correction rejection; observed [2026-03-29 120/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/bp-correct-ambiguous-reject/turn-1/events.jsonl`.
- `production/mt-bp-latest-then-correct`: turn 1: expected unchanged seed rows and latest blood-pressure answer; observed weights [] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest blood-pressure correction on 2026-03-30; observed weights [] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/mt-bp-latest-then-correct/turn-2/events.jsonl`.
- `production/mt-mixed-latest-then-correct`: turn 1: expected unchanged seed rows and latest mixed-domain answer; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest mixed-domain corrections on 2026-03-30; observed weights [2026-03-30 151.0 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
- `cli/bp-correct-existing`: turn 1: expected corrected 2026-03-29 blood-pressure row with no duplicate; observed [2026-03-29 121/77 pulse 63] Raw event reference: `<run-root>/cli/bp-correct-existing/turn-1/events.jsonl`.
- `cli/bp-correct-missing-reject`: turn 1: expected unchanged seed row and missing-date correction rejection; observed [2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/bp-correct-missing-reject/turn-1/events.jsonl`.
- `cli/bp-correct-ambiguous-reject`: turn 1: expected unchanged duplicate same-date rows and ambiguous correction rejection; observed [2026-03-29 120/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/bp-correct-ambiguous-reject/turn-1/events.jsonl`.
- `cli/mt-bp-latest-then-correct`: turn 1: expected unchanged seed rows and latest blood-pressure answer; observed weights [] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest blood-pressure correction on 2026-03-30; observed weights [] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/mt-bp-latest-then-correct/turn-2/events.jsonl`.
- `cli/mt-mixed-latest-then-correct`: turn 1: expected unchanged seed rows and latest mixed-domain answer; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest mixed-domain corrections on 2026-03-30; observed weights [2026-03-30 151.0 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/mt-mixed-latest-then-correct/turn-2/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands with the configured Go cache mode`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
