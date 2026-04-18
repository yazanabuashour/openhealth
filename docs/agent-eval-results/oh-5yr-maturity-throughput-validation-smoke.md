# oh-5yr Agent Eval Results

Date: maturity-throughput-validation-smoke

Harness: `codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `4`

Cache mode: `shared`

Cache prewarm seconds: `21.43`

Harness elapsed seconds: `11.32`

Effective parallel speedup: `2.75x`

Parallel efficiency: `0.69`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-maturity-throughput-validation-smoke.json`

Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.

## History Isolation

- Status: `passed`
- Single-turn ephemeral runs: `6`.
- Multi-turn persisted sessions: `0` sessions / `0` turns.
- The Codex desktop app was not used for eval prompts.
- New Codex session files referencing `<run-root>`: `0`.

## Results

| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |
| `production` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 4.42 | in 22398 / cached 18816 / out 142 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 0 | 1 | 4.69 | in 22376 / cached 18816 / out 180 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 0 | 1 | 4.88 | in 22414 / cached 19328 / out 190 | no | no | no | no | no | no |
| `production` | `bp-invalid-input` | pass | pass | pass | 0 | 1 | 6.20 | in 22387 / cached 18816 / out 250 | no | no | no | no | no | no |
| `production` | `bp-non-iso-date-reject` | pass | pass | pass | 0 | 1 | 4.41 | in 22432 / cached 18816 / out 204 | no | no | no | no | no | no |
| `production` | `mixed-invalid-direct-reject` | pass | pass | pass | 0 | 1 | 6.55 | in 22411 / cached 18816 / out 223 | no | no | no | no | no | no |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.20 |
| install_variant | 0.00 |
| warm_cache | 0.00 |
| seed_db | 0.04 |
| agent_run | 31.15 |
| parse_metrics | 0.00 |
| verify | 0.00 |
| total_job_time | 31.43 |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-maturity-throughput-validation-smoke.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `ambiguous-short-date` | same_pass | +0 | +0 | -0.50 | +167 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | +0 | +0 | -0.58 | +167 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | same_pass | +0 | +0 | -2.99 | -345 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-invalid-input` | same_pass | +0 | +0 | +1.87 | +167 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-non-iso-date-reject` | same_pass | +0 | +0 | -0.55 | +167 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `mixed-invalid-direct-reject` | same_pass | +0 | +0 | +1.77 | +167 | same_no | same_no | same_no | same_no | same_no | same_no |

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
| `candidate_passes_all_scenarios` | pass | 6/6 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `validation_scenarios_are_final_answer_only` | pass | validation scenarios used no tools, no command executions, and at most one assistant answer |
| `total_tools_less_than_or_equal_cli` | fail | production tools 0 vs cli tools 0 |
| `minimum_scenarios_at_or_below_cli` | fail | 0 scenarios at or below CLI tools; required 5 of 6 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | fail | missing cli scenarios: ambiguous-short-date, bp-invalid-input, bp-non-iso-date-reject, invalid-input, mixed-invalid-direct-reject, non-iso-date-reject |
| `non_cached_token_majority` | fail | 0 scenarios with lower non-cached input tokens; required 1 of 0; missing usage: none |
| `non_cached_token_total_less_than_or_equal_cli` | fail | production non-cached input tokens 0 vs cli 0; missing usage: none |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `ambiguous-short-date` | 0 | 0 | n/a |
| `invalid-input` | 0 | 0 | n/a |
| `non-iso-date-reject` | 0 | 0 | n/a |
| `bp-invalid-input` | 0 | 0 | n/a |
| `bp-non-iso-date-reject` | 0 | 0 | n/a |
| `mixed-invalid-direct-reject` | 0 | 0 | n/a |

## Scenario Notes

- `production/ambiguous-short-date`: turn 1: expected no write and a year clarification; observed [] Raw event reference: `<run-root>/production/ambiguous-short-date/turn-1/events.jsonl`.
- `production/invalid-input`: turn 1: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/production/invalid-input/turn-1/events.jsonl`.
- `production/non-iso-date-reject`: turn 1: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/production/non-iso-date-reject/turn-1/events.jsonl`.
- `production/bp-invalid-input`: turn 1: expected no write and an invalid blood-pressure rejection; observed [] Raw event reference: `<run-root>/production/bp-invalid-input/turn-1/events.jsonl`.
- `production/bp-non-iso-date-reject`: turn 1: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/production/bp-non-iso-date-reject/turn-1/events.jsonl`.
- `production/mixed-invalid-direct-reject`: turn 1: expected no mixed-domain writes and a direct invalid input rejection; observed weights [] and blood pressures [] Raw event reference: `<run-root>/production/mixed-invalid-direct-reject/turn-1/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands with the configured Go cache mode`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
