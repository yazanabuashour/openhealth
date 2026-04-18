# oh-5yr Agent Eval Results

Date: oh-23a-final

Harness: `codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `4`

Cache mode: `shared`

Cache prewarm seconds: `34.84`

Harness elapsed seconds: `169.17`

Effective parallel speedup: `3.80x`

Parallel efficiency: `0.95`

Reduced JSON artifact: `docs/agent-eval-results/archive/oh-5yr-oh-23a-final.json`

Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.

## History Isolation

- Status: `passed`
- Single-turn ephemeral runs: `46`.
- Multi-turn persisted sessions: `6` sessions / `12` turns.
- The Codex desktop app was not used for eval prompts.
- New Codex session files referencing `<run-root>`: `6`.

## Results

| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | pass | pass | pass | 1 | 2 | 13.13 | in 45621 / cached 41728 / out 722 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 1 | 1 | 13.42 | in 45359 / cached 41216 / out 466 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 1 | 3 | 24.20 | in 68388 / cached 64128 / out 715 | no | no | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 1 | 1 | 13.11 | in 68153 / cached 63616 / out 428 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 1 | 1 | 9.48 | in 45254 / cached 41216 / out 449 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 1 | 1 | 9.02 | in 67841 / cached 64128 / out 339 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 1 | 1 | 10.75 | in 67826 / cached 63616 / out 311 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 5.25 | in 22398 / cached 18816 / out 257 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 0 | 1 | 4.77 | in 22376 / cached 18816 / out 276 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 0 | 1 | 5.37 | in 22414 / cached 18816 / out 179 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 1 | 2 | 12.59 | in 45397 / cached 41216 / out 517 | no | no | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 1 | 1 | 10.74 | in 45110 / cached 41216 / out 318 | no | no | no | no | no | no |
| `production` | `bp-history-limit-two` | pass | pass | pass | 1 | 1 | 10.77 | in 45202 / cached 41216 / out 375 | no | no | no | no | no | no |
| `production` | `bp-bounded-range` | pass | pass | pass | 1 | 2 | 11.77 | in 45338 / cached 41216 / out 434 | no | no | no | no | no | no |
| `production` | `bp-bounded-range-natural` | pass | pass | pass | 1 | 2 | 9.23 | in 45234 / cached 41216 / out 409 | no | no | no | no | no | no |
| `production` | `bp-invalid-input` | pass | pass | pass | 0 | 1 | 7.19 | in 22387 / cached 18816 / out 141 | no | no | no | no | no | no |
| `production` | `bp-non-iso-date-reject` | pass | pass | pass | 0 | 1 | 7.58 | in 22432 / cached 18816 / out 260 | no | no | no | no | no | no |
| `production` | `bp-correct-existing` | pass | pass | pass | 2 | 4 | 15.51 | in 92148 / cached 87552 / out 850 | no | no | no | no | no | no |
| `production` | `bp-correct-missing-reject` | pass | pass | pass | 1 | 1 | 9.56 | in 45289 / cached 41216 / out 391 | no | no | no | no | no | no |
| `production` | `bp-correct-ambiguous-reject` | pass | pass | pass | 1 | 2 | 11.80 | in 45431 / cached 41216 / out 527 | no | no | no | no | no | no |
| `production` | `mixed-add-latest` | pass | pass | pass | 4 | 4 | 24.55 | in 116722 / cached 112000 / out 1386 | no | no | no | no | no | no |
| `production` | `mixed-bounded-range` | pass | pass | pass | 2 | 4 | 13.52 | in 91154 / cached 86528 / out 610 | no | no | no | no | no | no |
| `production` | `mixed-invalid-direct-reject` | pass | pass | pass | 0 | 1 | 6.13 | in 22411 / cached 18816 / out 183 | no | no | no | no | no | no |
| `production` | `mt-weight-clarify-then-add` | pass | pass | pass | 1 | 3 | 16.78 | in 90642 / cached 82944 / out 829 | no | no | no | no | no | no |
| `production` | `mt-mixed-latest-then-correct` | pass | pass | pass | 4 | 4 | 19.95 | in 138223 / cached 128768 / out 1717 | no | no | no | no | no | no |
| `production` | `mt-bp-latest-then-correct` | pass | pass | pass | 2 | 4 | 24.48 | in 136494 / cached 127744 / out 1261 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 4 | 5 | 15.56 | in 116171 / cached 106368 / out 935 | no | no | no | no | yes | no |
| `cli` | `repeat-add` | pass | pass | pass | 4 | 3 | 16.87 | in 116030 / cached 107392 / out 990 | no | no | no | no | yes | no |
| `cli` | `update-existing` | pass | pass | pass | 6 | 6 | 24.22 | in 183272 / cached 172672 / out 1415 | no | yes | no | no | yes | no |
| `cli` | `bounded-range` | pass | pass | pass | 2 | 4 | 11.69 | in 91820 / cached 86528 / out 527 | no | no | no | no | yes | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 3 | 11.80 | in 68271 / cached 63616 / out 436 | no | no | no | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 2 | 3 | 12.35 | in 68261 / cached 63616 / out 455 | no | no | no | no | yes | no |
| `cli` | `history-limit-two` | pass | pass | pass | 2 | 4 | 10.49 | in 91574 / cached 87040 / out 518 | no | no | no | no | yes | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 4.47 | in 22071 / cached 18816 / out 128 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 0 | 1 | 4.96 | in 22049 / cached 18816 / out 234 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 0 | 1 | 5.02 | in 22087 / cached 18816 / out 111 | no | no | no | no | no | no |
| `cli` | `bp-add-two` | pass | pass | pass | 2 | 4 | 14.80 | in 92059 / cached 86528 / out 696 | no | no | no | no | yes | no |
| `cli` | `bp-latest-only` | pass | pass | pass | 2 | 3 | 9.19 | in 68310 / cached 63616 / out 466 | no | no | no | no | yes | no |
| `cli` | `bp-history-limit-two` | pass | pass | pass | 2 | 3 | 11.44 | in 68223 / cached 63616 / out 378 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range` | pass | pass | pass | 2 | 3 | 9.46 | in 68396 / cached 63616 / out 420 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range-natural` | pass | pass | pass | 2 | 3 | 9.09 | in 68384 / cached 63616 / out 468 | no | no | no | no | yes | no |
| `cli` | `bp-invalid-input` | pass | pass | pass | 0 | 1 | 4.11 | in 22060 / cached 18816 / out 88 | no | no | no | no | no | no |
| `cli` | `bp-non-iso-date-reject` | pass | pass | pass | 0 | 1 | 6.64 | in 22105 / cached 18816 / out 170 | no | no | no | no | no | no |
| `cli` | `bp-correct-existing` | pass | pass | pass | 2 | 3 | 13.86 | in 68393 / cached 60544 / out 471 | no | no | no | no | yes | no |
| `cli` | `bp-correct-missing-reject` | pass | pass | pass | 2 | 3 | 11.69 | in 68714 / cached 63616 / out 646 | no | no | no | no | yes | no |
| `cli` | `bp-correct-ambiguous-reject` | pass | pass | pass | 2 | 4 | 14.81 | in 92247 / cached 87040 / out 814 | no | no | no | no | yes | no |
| `cli` | `mixed-add-latest` | pass | pass | pass | 4 | 4 | 22.52 | in 139582 / cached 133376 / out 981 | no | no | no | no | yes | no |
| `cli` | `mixed-bounded-range` | pass | pass | pass | 3 | 3 | 10.34 | in 68540 / cached 63616 / out 633 | no | no | no | no | yes | no |
| `cli` | `mixed-invalid-direct-reject` | pass | pass | pass | 0 | 1 | 4.69 | in 22084 / cached 18816 / out 239 | no | no | no | no | no | no |
| `cli` | `mt-weight-clarify-then-add` | pass | pass | pass | 2 | 4 | 13.13 | in 113053 / cached 104832 / out 707 | no | no | no | no | yes | no |
| `cli` | `mt-mixed-latest-then-correct` | pass | pass | pass | 9 | 9 | 26.21 | in 307657 / cached 290688 / out 2769 | no | no | no | no | yes | no |
| `cli` | `mt-bp-latest-then-correct` | pass | pass | pass | 4 | 7 | 22.20 | in 254701 / cached 242816 / out 1438 | no | no | no | no | yes | no |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 1.23 |
| install_variant | 0.01 |
| warm_cache | 0.00 |
| seed_db | 0.20 |
| agent_run | 642.26 |
| parse_metrics | 0.00 |
| verify | 0.01 |
| total_job_time | 643.97 |

## Comparison

Baseline: `docs/agent-eval-results/archive/oh-5yr-oh-967-smoke.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | -3 | -3 | -20.08 | -2750 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `repeat-add` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `update-existing` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bounded-range-natural` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `latest-only` | same_pass | -1 | -3 | -18.98 | -2010 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `history-limit-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `ambiguous-short-date` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `invalid-input` | same_pass | -2 | -2 | -4.07 | -1388 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-add-two` | same_pass | -2 | -3 | -20.95 | -2345 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-latest-only` | same_pass | -2 | -2 | -8.99 | -1387 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-history-limit-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-bounded-range-natural` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-invalid-input` | same_pass | -2 | -2 | -3.95 | -4042 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-non-iso-date-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-correct-existing` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-correct-missing-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-correct-ambiguous-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mixed-add-latest` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mixed-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mixed-invalid-direct-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mt-weight-clarify-then-add` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mt-mixed-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mt-bp-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `add-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `repeat-add` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `update-existing` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bounded-range-natural` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `latest-only` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `history-limit-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `ambiguous-short-date` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `invalid-input` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `non-iso-date-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-add-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-latest-only` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-history-limit-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-bounded-range-natural` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-invalid-input` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-non-iso-date-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-correct-existing` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-correct-missing-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-correct-ambiguous-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mixed-add-latest` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mixed-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mixed-invalid-direct-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mt-weight-clarify-then-add` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mt-mixed-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mt-bp-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |

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
| `candidate_passes_all_scenarios` | pass | 26/26 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `validation_scenarios_are_final_answer_only` | pass | validation scenarios used no tools, no command executions, and at most one assistant answer |
| `total_tools_less_than_or_equal_cli` | pass | production tools 29 vs cli tools 60 |
| `minimum_scenarios_at_or_below_cli` | pass | 26 scenarios at or below CLI tools; required 21 of 26 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |
| `non_cached_token_majority` | pass | 20 scenarios with lower non-cached input tokens; required 14 of 26; missing usage: none |
| `non_cached_token_total_less_than_or_equal_cli` | pass | production non-cached input tokens 118652 vs cli 158466; missing usage: none |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 1 | 4 | -3 |
| `repeat-add` | 1 | 4 | -3 |
| `update-existing` | 1 | 6 | -5 |
| `bounded-range` | 1 | 2 | -1 |
| `bounded-range-natural` | 1 | 2 | -1 |
| `latest-only` | 1 | 2 | -1 |
| `history-limit-two` | 1 | 2 | -1 |
| `ambiguous-short-date` | 0 | 0 | +0 |
| `invalid-input` | 0 | 0 | +0 |
| `non-iso-date-reject` | 0 | 0 | +0 |
| `bp-add-two` | 1 | 2 | -1 |
| `bp-latest-only` | 1 | 2 | -1 |
| `bp-history-limit-two` | 1 | 2 | -1 |
| `bp-bounded-range` | 1 | 2 | -1 |
| `bp-bounded-range-natural` | 1 | 2 | -1 |
| `bp-invalid-input` | 0 | 0 | +0 |
| `bp-non-iso-date-reject` | 0 | 0 | +0 |
| `bp-correct-existing` | 2 | 2 | +0 |
| `bp-correct-missing-reject` | 1 | 2 | -1 |
| `bp-correct-ambiguous-reject` | 1 | 2 | -1 |
| `mixed-add-latest` | 4 | 4 | +0 |
| `mixed-bounded-range` | 2 | 3 | -1 |
| `mixed-invalid-direct-reject` | 0 | 0 | +0 |
| `mt-weight-clarify-then-add` | 1 | 2 | -1 |
| `mt-mixed-latest-then-correct` | 4 | 9 | -5 |
| `mt-bp-latest-then-correct` | 2 | 4 | -2 |

## Metric Evidence

- `cli/add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/repeat-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/update-existing` broad repo search: `/bin/zsh -lc 'rg -n "weight add|same-date|upsert|replace|duplicate|ON CONFLICT|INSERT OR" cmd internal .'`.
- `cli/update-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight --help'`.
- `cli/bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 2'`.
- `cli/bp-add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64 && go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 118 --diastolic 76 && go run ./cmd/openhealth blood-pressure list --limit 25'`.
- `cli/bp-latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/bp-history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 2'`.
- `cli/bp-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-correct-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure correct --date 2026-03-29 --systolic 121 --diastolic 77 --pulse 63 && go run ./cmd/openhealth blood-pressure list --limit 25'`.
- `cli/bp-correct-missing-reject` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure correct --date 2026-03-31 --systolic 121 --diastolic 77'`.
- `cli/bp-correct-ambiguous-reject` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/mixed-add-latest` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-31 --value 150.8 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-31 --systolic 119 --diastolic 77 --pulse 62'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1 && go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/mixed-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/mt-weight-clarify-then-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`.
- `cli/mt-mixed-latest-then-correct` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight correct --date 2026-03-30 --value 151.0 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure correct --date 2026-03-30 --systolic 117 --diastolic 75 --pulse 63'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight --help'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.0 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/mt-bp-latest-then-correct` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure correct --date 2026-03-30 --systolic 117 --diastolic 75 --pulse 63'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.

## Turn Details

- `production/mt-weight-clarify-then-add` turn 1: exit `0`, tools `0`, assistant calls `1`, wall `7.11`, raw `<run-root>/production/mt-weight-clarify-then-add/turn-1/events.jsonl`.
- `production/mt-weight-clarify-then-add` turn 2: exit `0`, tools `1`, assistant calls `2`, wall `9.67`, raw `<run-root>/production/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `production/mt-mixed-latest-then-correct` turn 1: exit `0`, tools `2`, assistant calls `2`, wall `9.78`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-1/events.jsonl`.
- `production/mt-mixed-latest-then-correct` turn 2: exit `0`, tools `2`, assistant calls `2`, wall `10.17`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
- `production/mt-bp-latest-then-correct` turn 1: exit `0`, tools `1`, assistant calls `2`, wall `10.92`, raw `<run-root>/production/mt-bp-latest-then-correct/turn-1/events.jsonl`.
- `production/mt-bp-latest-then-correct` turn 2: exit `0`, tools `1`, assistant calls `2`, wall `13.56`, raw `<run-root>/production/mt-bp-latest-then-correct/turn-2/events.jsonl`.
- `cli/mt-weight-clarify-then-add` turn 1: exit `0`, tools `0`, assistant calls `1`, wall `4.38`, raw `<run-root>/cli/mt-weight-clarify-then-add/turn-1/events.jsonl`.
- `cli/mt-weight-clarify-then-add` turn 2: exit `0`, tools `2`, assistant calls `3`, wall `8.75`, raw `<run-root>/cli/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `cli/mt-mixed-latest-then-correct` turn 1: exit `0`, tools `3`, assistant calls `4`, wall `10.54`, raw `<run-root>/cli/mt-mixed-latest-then-correct/turn-1/events.jsonl`.
- `cli/mt-mixed-latest-then-correct` turn 2: exit `0`, tools `6`, assistant calls `5`, wall `15.67`, raw `<run-root>/cli/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
- `cli/mt-bp-latest-then-correct` turn 1: exit `0`, tools `2`, assistant calls `4`, wall `14.30`, raw `<run-root>/cli/mt-bp-latest-then-correct/turn-1/events.jsonl`.
- `cli/mt-bp-latest-then-correct` turn 2: exit `0`, tools `2`, assistant calls `3`, wall `7.90`, raw `<run-root>/cli/mt-bp-latest-then-correct/turn-2/events.jsonl`.

## Scenario Notes

- `production/add-two`: turn 1: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/add-two/turn-1/events.jsonl`.
- `production/repeat-add`: turn 1: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/repeat-add/turn-1/events.jsonl`.
- `production/update-existing`: turn 1: expected one updated row; observed [2026-03-29 151.6 lb] Raw event reference: `<run-root>/production/update-existing/turn-1/events.jsonl`.
- `production/bounded-range`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/bounded-range/turn-1/events.jsonl`.
- `production/bounded-range-natural`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/bounded-range-natural/turn-1/events.jsonl`.
- `production/latest-only`: turn 1: expected unchanged seed rows and assistant output limited to latest row 2026-03-30; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/latest-only/turn-1/events.jsonl`.
- `production/history-limit-two`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb, 2026-03-27 154.1 lb] Raw event reference: `<run-root>/production/history-limit-two/turn-1/events.jsonl`.
- `production/ambiguous-short-date`: turn 1: expected no write and a year clarification; observed [] Raw event reference: `<run-root>/production/ambiguous-short-date/turn-1/events.jsonl`.
- `production/invalid-input`: turn 1: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/production/invalid-input/turn-1/events.jsonl`.
- `production/non-iso-date-reject`: turn 1: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/production/non-iso-date-reject/turn-1/events.jsonl`.
- `production/bp-add-two`: turn 1: expected exactly two newest-first blood-pressure rows; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/bp-add-two/turn-1/events.jsonl`.
- `production/bp-latest-only`: turn 1: expected unchanged seed rows and assistant output limited to latest blood-pressure row 2026-03-30; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/bp-latest-only/turn-1/events.jsonl`.
- `production/bp-history-limit-two`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80, 2026-03-27 126/82] Raw event reference: `<run-root>/production/bp-history-limit-two/turn-1/events.jsonl`.
- `production/bp-bounded-range`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/bp-bounded-range/turn-1/events.jsonl`.
- `production/bp-bounded-range-natural`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/bp-bounded-range-natural/turn-1/events.jsonl`.
- `production/bp-invalid-input`: turn 1: expected no write and an invalid blood-pressure rejection; observed [] Raw event reference: `<run-root>/production/bp-invalid-input/turn-1/events.jsonl`.
- `production/bp-non-iso-date-reject`: turn 1: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/production/bp-non-iso-date-reject/turn-1/events.jsonl`.
- `production/bp-correct-existing`: turn 1: expected corrected 2026-03-29 blood-pressure row with no duplicate; observed [2026-03-29 121/77 pulse 63] Raw event reference: `<run-root>/production/bp-correct-existing/turn-1/events.jsonl`.
- `production/bp-correct-missing-reject`: turn 1: expected unchanged seed row and missing-date correction rejection; observed [2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/bp-correct-missing-reject/turn-1/events.jsonl`.
- `production/bp-correct-ambiguous-reject`: turn 1: expected unchanged duplicate same-date rows and ambiguous correction rejection; observed [2026-03-29 120/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/bp-correct-ambiguous-reject/turn-1/events.jsonl`.
- `production/mixed-add-latest`: turn 1: expected latest weight and blood-pressure rows for 2026-03-31; observed weights [2026-03-31 150.8 lb] and blood pressures [2026-03-31 119/77 pulse 62] Raw event reference: `<run-root>/production/mixed-add-latest/turn-1/events.jsonl`.
- `production/mixed-bounded-range`: turn 1: expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/mixed-bounded-range/turn-1/events.jsonl`.
- `production/mixed-invalid-direct-reject`: turn 1: expected no mixed-domain writes and a direct invalid input rejection; observed weights [] and blood pressures [] Raw event reference: `<run-root>/production/mixed-invalid-direct-reject/turn-1/events.jsonl`.
- `production/mt-weight-clarify-then-add`: turn 1: expected no first-turn writes and a year clarification; observed weights [] and blood pressures []; turn 2: expected second-turn weight write after year clarification with no blood-pressure writes; observed weights [2026-03-29 152.2 lb] and blood pressures [] Raw event reference: `<run-root>/production/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `production/mt-mixed-latest-then-correct`: turn 1: expected unchanged seed rows and latest mixed-domain answer; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest mixed-domain corrections on 2026-03-30; observed weights [2026-03-30 151.0 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
- `production/mt-bp-latest-then-correct`: turn 1: expected unchanged seed rows and latest blood-pressure answer; observed weights [] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest blood-pressure correction on 2026-03-30; observed weights [] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/mt-bp-latest-then-correct/turn-2/events.jsonl`.
- `cli/add-two`: turn 1: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/cli/add-two/turn-1/events.jsonl`.
- `cli/repeat-add`: turn 1: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/cli/repeat-add/turn-1/events.jsonl`.
- `cli/update-existing`: turn 1: expected one updated row; observed [2026-03-29 151.6 lb] Raw event reference: `<run-root>/cli/update-existing/turn-1/events.jsonl`.
- `cli/bounded-range`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/cli/bounded-range/turn-1/events.jsonl`.
- `cli/bounded-range-natural`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/cli/bounded-range-natural/turn-1/events.jsonl`.
- `cli/latest-only`: turn 1: expected unchanged seed rows and assistant output limited to latest row 2026-03-30; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/cli/latest-only/turn-1/events.jsonl`.
- `cli/history-limit-two`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb, 2026-03-27 154.1 lb] Raw event reference: `<run-root>/cli/history-limit-two/turn-1/events.jsonl`.
- `cli/ambiguous-short-date`: turn 1: expected no write and a year clarification; observed [] Raw event reference: `<run-root>/cli/ambiguous-short-date/turn-1/events.jsonl`.
- `cli/invalid-input`: turn 1: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/cli/invalid-input/turn-1/events.jsonl`.
- `cli/non-iso-date-reject`: turn 1: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/cli/non-iso-date-reject/turn-1/events.jsonl`.
- `cli/bp-add-two`: turn 1: expected exactly two newest-first blood-pressure rows; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/bp-add-two/turn-1/events.jsonl`.
- `cli/bp-latest-only`: turn 1: expected unchanged seed rows and assistant output limited to latest blood-pressure row 2026-03-30; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/bp-latest-only/turn-1/events.jsonl`.
- `cli/bp-history-limit-two`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80, 2026-03-27 126/82] Raw event reference: `<run-root>/cli/bp-history-limit-two/turn-1/events.jsonl`.
- `cli/bp-bounded-range`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/bp-bounded-range/turn-1/events.jsonl`.
- `cli/bp-bounded-range-natural`: turn 1: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/bp-bounded-range-natural/turn-1/events.jsonl`.
- `cli/bp-invalid-input`: turn 1: expected no write and an invalid blood-pressure rejection; observed [] Raw event reference: `<run-root>/cli/bp-invalid-input/turn-1/events.jsonl`.
- `cli/bp-non-iso-date-reject`: turn 1: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/cli/bp-non-iso-date-reject/turn-1/events.jsonl`.
- `cli/bp-correct-existing`: turn 1: expected corrected 2026-03-29 blood-pressure row with no duplicate; observed [2026-03-29 121/77 pulse 63] Raw event reference: `<run-root>/cli/bp-correct-existing/turn-1/events.jsonl`.
- `cli/bp-correct-missing-reject`: turn 1: expected unchanged seed row and missing-date correction rejection; observed [2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/bp-correct-missing-reject/turn-1/events.jsonl`.
- `cli/bp-correct-ambiguous-reject`: turn 1: expected unchanged duplicate same-date rows and ambiguous correction rejection; observed [2026-03-29 120/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/bp-correct-ambiguous-reject/turn-1/events.jsonl`.
- `cli/mixed-add-latest`: turn 1: expected latest weight and blood-pressure rows for 2026-03-31; observed weights [2026-03-31 150.8 lb] and blood pressures [2026-03-31 119/77 pulse 62] Raw event reference: `<run-root>/cli/mixed-add-latest/turn-1/events.jsonl`.
- `cli/mixed-bounded-range`: turn 1: expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/mixed-bounded-range/turn-1/events.jsonl`.
- `cli/mixed-invalid-direct-reject`: turn 1: expected no mixed-domain writes and a direct invalid input rejection; observed weights [] and blood pressures [] Raw event reference: `<run-root>/cli/mixed-invalid-direct-reject/turn-1/events.jsonl`.
- `cli/mt-weight-clarify-then-add`: turn 1: expected no first-turn writes and a year clarification; observed weights [] and blood pressures []; turn 2: expected second-turn weight write after year clarification with no blood-pressure writes; observed weights [2026-03-29 152.2 lb] and blood pressures [] Raw event reference: `<run-root>/cli/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `cli/mt-mixed-latest-then-correct`: turn 1: expected unchanged seed rows and latest mixed-domain answer; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest mixed-domain corrections on 2026-03-30; observed weights [2026-03-30 151.0 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
- `cli/mt-bp-latest-then-correct`: turn 1: expected unchanged seed rows and latest blood-pressure answer; observed weights [] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest blood-pressure correction on 2026-03-30; observed weights [] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/mt-bp-latest-then-correct/turn-2/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands with the configured Go cache mode`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
