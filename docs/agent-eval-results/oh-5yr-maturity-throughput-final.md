# oh-5yr Agent Eval Results

Date: maturity-throughput-final

Harness: `codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `4`

Cache mode: `shared`

Cache prewarm seconds: `22.40`

Harness elapsed seconds: `157.26`

Effective parallel speedup: `3.39x`

Parallel efficiency: `0.85`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-maturity-throughput-final.json`

Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.

## History Isolation

- Status: `passed`
- Single-turn ephemeral runs: `40`.
- Multi-turn persisted sessions: `4` sessions / `8` turns.
- The Codex desktop app was not used for eval prompts.
- New Codex session files referencing `<run-root>`: `4`.

## Results

| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | pass | pass | pass | 1 | 2 | 12.79 | in 45496 / cached 41216 / out 764 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 1 | 2 | 12.75 | in 45397 / cached 41216 / out 549 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 2 | 4 | 13.03 | in 91573 / cached 86528 / out 711 | no | no | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 2 | 3 | 14.71 | in 69720 / cached 64640 / out 636 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 1 | 3 | 11.54 | in 67861 / cached 63616 / out 331 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 1 | 2 | 12.31 | in 45121 / cached 41216 / out 345 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 1 | 1 | 8.84 | in 45159 / cached 41216 / out 398 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 5.42 | in 22391 / cached 18816 / out 179 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 0 | 1 | 4.79 | in 22369 / cached 18816 / out 247 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 0 | 1 | 6.92 | in 22407 / cached 18816 / out 213 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 2 | 4 | 17.71 | in 92422 / cached 88576 / out 998 | no | no | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 1 | 1 | 9.54 | in 67982 / cached 63616 / out 382 | no | no | no | no | no | no |
| `production` | `bp-history-limit-two` | pass | pass | pass | 1 | 3 | 11.02 | in 68081 / cached 63616 / out 443 | no | no | no | no | no | no |
| `production` | `bp-bounded-range` | pass | pass | pass | 1 | 2 | 9.49 | in 45408 / cached 41216 / out 493 | no | no | no | no | no | no |
| `production` | `bp-bounded-range-natural` | pass | pass | pass | 1 | 2 | 13.42 | in 45197 / cached 41216 / out 391 | no | no | no | no | no | no |
| `production` | `bp-invalid-input` | pass | pass | pass | 0 | 1 | 4.74 | in 22380 / cached 18816 / out 219 | no | no | no | no | no | no |
| `production` | `bp-non-iso-date-reject` | pass | pass | pass | 0 | 1 | 5.60 | in 22425 / cached 18816 / out 233 | no | no | no | no | no | no |
| `production` | `mixed-add-latest` | pass | pass | pass | 6 | 4 | 15.94 | in 92857 / cached 88064 / out 1194 | no | no | no | no | no | no |
| `production` | `mixed-bounded-range` | pass | pass | pass | 2 | 3 | 11.28 | in 68821 / cached 64128 / out 785 | no | no | no | no | no | no |
| `production` | `mixed-invalid-direct-reject` | pass | pass | pass | 0 | 1 | 5.23 | in 22404 / cached 18816 / out 288 | no | no | no | no | no | no |
| `production` | `mt-weight-clarify-then-add` | pass | pass | pass | 1 | 4 | 17.46 | in 113618 / cached 105344 / out 804 | no | no | no | no | no | no |
| `production` | `mt-mixed-latest-then-correct` | pass | pass | pass | 4 | 5 | 22.72 | in 184974 / cached 170496 / out 1921 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 4 | 3 | 17.75 | in 115147 / cached 109952 / out 826 | no | no | no | no | yes | no |
| `cli` | `repeat-add` | pass | pass | pass | 4 | 4 | 15.14 | in 115163 / cached 106880 / out 882 | no | no | no | no | yes | no |
| `cli` | `update-existing` | pass | pass | pass | 7 | 4 | 20.53 | in 214182 / cached 202752 / out 1283 | no | yes | no | no | yes | no |
| `cli` | `bounded-range` | pass | pass | pass | 2 | 3 | 12.14 | in 69035 / cached 64128 / out 557 | no | no | no | no | yes | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 3 | 9.52 | in 68038 / cached 59520 / out 518 | no | no | no | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 2 | 4 | 14.07 | in 91146 / cached 86528 / out 532 | no | no | no | no | yes | no |
| `cli` | `history-limit-two` | pass | pass | pass | 2 | 3 | 10.70 | in 67057 / cached 63104 / out 497 | no | no | no | no | yes | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 5.34 | in 22044 / cached 18816 / out 136 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 0 | 1 | 4.77 | in 22022 / cached 18816 / out 131 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 0 | 1 | 5.54 | in 22060 / cached 18816 / out 304 | no | no | no | no | no | no |
| `cli` | `bp-add-two` | pass | pass | pass | 4 | 4 | 14.18 | in 91862 / cached 87040 / out 826 | no | no | no | no | yes | no |
| `cli` | `bp-latest-only` | pass | pass | pass | 2 | 4 | 9.62 | in 91213 / cached 86528 / out 490 | no | no | no | no | yes | no |
| `cli` | `bp-history-limit-two` | pass | pass | pass | 2 | 3 | 11.41 | in 68096 / cached 63616 / out 458 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range` | pass | pass | pass | 2 | 3 | 12.78 | in 68188 / cached 63616 / out 507 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range-natural` | pass | pass | pass | 2 | 3 | 14.98 | in 68085 / cached 63616 / out 511 | no | no | no | no | yes | no |
| `cli` | `bp-invalid-input` | pass | pass | pass | 0 | 1 | 4.51 | in 22033 / cached 18816 / out 179 | no | no | no | no | no | no |
| `cli` | `bp-non-iso-date-reject` | pass | pass | pass | 0 | 1 | 4.06 | in 22078 / cached 18816 / out 130 | no | no | no | no | no | no |
| `cli` | `mixed-add-latest` | pass | pass | pass | 5 | 4 | 18.35 | in 92155 / cached 87040 / out 997 | no | no | no | no | yes | no |
| `cli` | `mixed-bounded-range` | pass | pass | pass | 2 | 4 | 11.16 | in 91471 / cached 86528 / out 602 | no | no | no | no | yes | no |
| `cli` | `mixed-invalid-direct-reject` | pass | pass | pass | 0 | 1 | 4.13 | in 22057 / cached 18816 / out 162 | no | no | no | no | no | no |
| `cli` | `mt-weight-clarify-then-add` | pass | pass | pass | 2 | 4 | 23.63 | in 136235 / cached 127232 / out 802 | no | no | no | no | yes | no |
| `cli` | `mt-mixed-latest-then-correct` | fail | fail | pass | 13 | 10 | 41.51 | in 356551 / cached 343168 / out 3993 | no | no | no | no | yes | no |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 1.12 |
| install_variant | 0.00 |
| warm_cache | 0.00 |
| seed_db | 0.10 |
| agent_run | 533.07 |
| parse_metrics | 0.00 |
| verify | 0.01 |
| total_job_time | 534.57 |

## Comparison

Baseline: `docs/agent-eval-results/archive/oh-5yr-oh-967-final-r2.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | -1 | -2 | -22.35 | -6907 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `repeat-add` | same_pass | -1 | -2 | -16.50 | -6214 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `update-existing` | same_pass | +0 | -1 | -20.88 | -1054 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range` | same_pass | +0 | -1 | -16.96 | -1541 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range-natural` | same_pass | -1 | +0 | -12.45 | -545 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | -1 | -2 | -8.82 | -950 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `history-limit-two` | same_pass | -1 | -3 | -11.21 | -991 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `ambiguous-short-date` | same_pass | -1 | -1 | -3.54 | -338 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | -1 | -1 | -3.43 | -1399 | same_no | improved_to_no | improved_to_no | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | same_pass | -1 | -1 | -0.23 | -837 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | same_pass | +0 | +0 | -6.15 | -7022 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-latest-only` | same_pass | -1 | -3 | -10.86 | -5170 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-history-limit-two` | same_pass | -1 | -1 | -13.92 | -940 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range` | same_pass | -1 | -2 | -13.36 | -1693 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range-natural` | same_pass | -1 | -2 | -6.23 | -5775 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-invalid-input` | same_pass | -1 | -1 | -4.00 | -1405 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-non-iso-date-reject` | same_pass | -1 | -1 | -1.76 | -1442 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `mixed-add-latest` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mixed-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mixed-invalid-direct-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mt-weight-clarify-then-add` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `mt-mixed-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `add-two` | same_pass | +0 | -1 | -6.57 | -567 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `repeat-add` | same_pass | +0 | -2 | -14.88 | +3736 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `update-existing` | same_pass | +0 | -2 | -20.05 | -11821 | same_no | same_yes | improved_to_no | same_no | same_yes | same_no |
| `cli` | `bounded-range` | same_pass | +0 | -1 | -15.33 | +158 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bounded-range-natural` | same_pass | +0 | -2 | -11.10 | +3451 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `latest-only` | same_pass | +0 | +0 | -15.35 | +202 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `history-limit-two` | same_pass | +0 | -1 | -14.62 | -1682 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `ambiguous-short-date` | same_pass | -1 | -1 | -4.87 | -726 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | -1 | -1 | -6.59 | -165 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `non-iso-date-reject` | same_pass | -1 | -1 | -1.81 | -3254 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `bp-add-two` | same_pass | +0 | -1 | -23.82 | -3875 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-latest-only` | same_pass | +0 | +0 | -24.51 | +303 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-history-limit-two` | same_pass | +0 | -1 | -12.55 | +7 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range` | same_pass | -2 | -4 | -23.27 | -5680 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range-natural` | same_pass | +0 | +0 | -2.70 | +223 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-invalid-input` | same_pass | -8 | -5 | -26.19 | -21087 | improved_to_no | same_no | same_no | same_no | improved_to_no | same_no |
| `cli` | `bp-non-iso-date-reject` | same_pass | -2 | -1 | -4.84 | -924 | same_no | improved_to_no | improved_to_no | same_no | same_no | same_no |
| `cli` | `mixed-add-latest` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mixed-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mixed-invalid-direct-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mt-weight-clarify-then-add` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `mt-mixed-latest-then-correct` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |

## Production Stop-Loss

- Triggered: `no`
- Recommendation: `continue_production_hardening`

## Code-First CLI Comparison

- Candidate: `production`
- Baseline: `cli`
- Beats CLI: `yes`
- Recommendation: `prefer_openhealth_runner_for_routine_openhealth_operations`

| Criterion | Result | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | pass | 22/22 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `validation_scenarios_are_final_answer_only` | pass | validation scenarios used no tools, no command executions, and at most one assistant answer |
| `total_tools_less_than_or_equal_cli` | pass | production tools 28 vs cli tools 57 |
| `minimum_scenarios_at_or_below_cli` | pass | 21 scenarios at or below CLI tools; required 18 of 22 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |
| `non_cached_token_majority` | pass | 14 scenarios with lower non-cached input tokens; required 12 of 22; missing usage: none |
| `non_cached_token_total_less_than_or_equal_cli` | pass | production non-cached input tokens 105247 vs cli 121774; missing usage: none |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 1 | 4 | -3 |
| `repeat-add` | 1 | 4 | -3 |
| `update-existing` | 2 | 7 | -5 |
| `bounded-range` | 2 | 2 | +0 |
| `bounded-range-natural` | 1 | 2 | -1 |
| `latest-only` | 1 | 2 | -1 |
| `history-limit-two` | 1 | 2 | -1 |
| `ambiguous-short-date` | 0 | 0 | +0 |
| `invalid-input` | 0 | 0 | +0 |
| `non-iso-date-reject` | 0 | 0 | +0 |
| `bp-add-two` | 2 | 4 | -2 |
| `bp-latest-only` | 1 | 2 | -1 |
| `bp-history-limit-two` | 1 | 2 | -1 |
| `bp-bounded-range` | 1 | 2 | -1 |
| `bp-bounded-range-natural` | 1 | 2 | -1 |
| `bp-invalid-input` | 0 | 0 | +0 |
| `bp-non-iso-date-reject` | 0 | 0 | +0 |
| `mixed-add-latest` | 6 | 5 | +1 |
| `mixed-bounded-range` | 2 | 2 | +0 |
| `mixed-invalid-direct-reject` | 0 | 0 | +0 |
| `mt-weight-clarify-then-add` | 1 | 2 | -1 |
| `mt-mixed-latest-then-correct` | 4 | 13 | -9 |

## Metric Evidence

- `cli/add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/repeat-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/update-existing` broad repo search: `/bin/zsh -lc "rg -n \"weight (add|list|update|edit|set)|blood-pressure (add|list|update|edit|set)\" cmd . -g '"'!**/vendor/**'"'"`.
- `cli/update-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-29'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 151.6 --unit lb'`.
- `cli/bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight --help'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 2'`.
- `cli/bp-add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 118 --diastolic 76'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 25'`.
- `cli/bp-latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/bp-history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 2'`.
- `cli/bp-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/mixed-add-latest` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-31 --systolic 119 --diastolic 77 --pulse 62'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-31 --value 150.8 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/mixed-bounded-range` openhealth CLI: `/bin/zsh -lc "go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30 && printf '\\n---\\n' && go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30"`.
- `cli/mt-weight-clarify-then-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`.
- `cli/mt-mixed-latest-then-correct` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight --help'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure --help'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.0 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 117 --diastolic 75 --pulse 63'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.

## Turn Details

- `production/mt-weight-clarify-then-add` turn 1: exit `0`, tools `0`, assistant calls `1`, wall `5.29`, raw `<run-root>/production/mt-weight-clarify-then-add/turn-1/events.jsonl`.
- `production/mt-weight-clarify-then-add` turn 2: exit `0`, tools `1`, assistant calls `3`, wall `12.17`, raw `<run-root>/production/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `production/mt-mixed-latest-then-correct` turn 1: exit `0`, tools `2`, assistant calls `3`, wall `14.06`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-1/events.jsonl`.
- `production/mt-mixed-latest-then-correct` turn 2: exit `0`, tools `2`, assistant calls `2`, wall `8.66`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
- `cli/mt-weight-clarify-then-add` turn 1: exit `0`, tools `0`, assistant calls `1`, wall `4.71`, raw `<run-root>/cli/mt-weight-clarify-then-add/turn-1/events.jsonl`.
- `cli/mt-weight-clarify-then-add` turn 2: exit `0`, tools `2`, assistant calls `3`, wall `18.92`, raw `<run-root>/cli/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `cli/mt-mixed-latest-then-correct` turn 1: exit `0`, tools `5`, assistant calls `5`, wall `15.31`, raw `<run-root>/cli/mt-mixed-latest-then-correct/turn-1/events.jsonl`.
- `cli/mt-mixed-latest-then-correct` turn 2: exit `0`, tools `8`, assistant calls `5`, wall `26.20`, raw `<run-root>/cli/mt-mixed-latest-then-correct/turn-2/events.jsonl`.

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
- `production/mixed-add-latest`: turn 1: expected latest weight and blood-pressure rows for 2026-03-31; observed weights [2026-03-31 150.8 lb] and blood pressures [2026-03-31 119/77 pulse 62] Raw event reference: `<run-root>/production/mixed-add-latest/turn-1/events.jsonl`.
- `production/mixed-bounded-range`: turn 1: expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/mixed-bounded-range/turn-1/events.jsonl`.
- `production/mixed-invalid-direct-reject`: turn 1: expected no mixed-domain writes and a direct invalid input rejection; observed weights [] and blood pressures [] Raw event reference: `<run-root>/production/mixed-invalid-direct-reject/turn-1/events.jsonl`.
- `production/mt-weight-clarify-then-add`: turn 1: expected no first-turn write and a year clarification; observed weights []; turn 2: expected second-turn write after year clarification; observed weights [2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `production/mt-mixed-latest-then-correct`: turn 1: expected unchanged seed rows and latest mixed-domain answer; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest mixed-domain corrections on 2026-03-30; observed weights [2026-03-30 151.0 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.
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
- `cli/mixed-add-latest`: turn 1: expected latest weight and blood-pressure rows for 2026-03-31; observed weights [2026-03-31 150.8 lb] and blood pressures [2026-03-31 119/77 pulse 62] Raw event reference: `<run-root>/cli/mixed-add-latest/turn-1/events.jsonl`.
- `cli/mixed-bounded-range`: turn 1: expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/mixed-bounded-range/turn-1/events.jsonl`.
- `cli/mixed-invalid-direct-reject`: turn 1: expected no mixed-domain writes and a direct invalid input rejection; observed weights [] and blood pressures [] Raw event reference: `<run-root>/cli/mixed-invalid-direct-reject/turn-1/events.jsonl`.
- `cli/mt-weight-clarify-then-add`: turn 1: expected no first-turn write and a year clarification; observed weights []; turn 2: expected second-turn write after year clarification; observed weights [2026-03-29 152.2 lb] Raw event reference: `<run-root>/cli/mt-weight-clarify-then-add/turn-2/events.jsonl`.
- `cli/mt-mixed-latest-then-correct`: turn 1: expected unchanged seed rows and latest mixed-domain answer; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest mixed-domain corrections on 2026-03-30; observed weights [2026-03-30 151.0 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/mt-mixed-latest-then-correct/turn-2/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands with the configured Go cache mode`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
