# oh-5yr Agent Eval Results

Date: agentops-production-expanded-r3

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-agentops-production-expanded-r3.json`

Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.

## History Isolation

- Status: `review`
- Every agent run used `codex exec --ephemeral` from `<run-root>/<variant>/<scenario>/repo`.
- The Codex desktop app was not used for eval prompts.
- New Codex session files referencing `<run-root>`: `1`.
- Limitation: A session-file count changed while evals ran; this may be from another Codex process, because the harness uses --ephemeral and a throwaway cwd.

## Results

| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | pass | pass | pass | 2 | 3 | 20.77 | in 72537 / cached 66688 / out 917 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 2 | 3 | 20.53 | in 72358 / cached 68736 / out 825 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 2 | 4 | 19.06 | in 123094 / cached 119168 / out 1187 | no | no | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 2 | 4 | 19.82 | in 97612 / cached 91136 / out 928 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 2 | 4 | 21.83 | in 122748 / cached 116096 / out 1005 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 2 | 3 | 19.06 | in 72057 / cached 68736 / out 700 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 2 | 4 | 20.11 | in 122264 / cached 115584 / out 882 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 4.32 | in 22816 / cached 18816 / out 223 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 1 | 2 | 8.25 | in 47187 / cached 44288 / out 378 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 1 | 2 | 9.91 | in 47298 / cached 41728 / out 453 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 8 | 6 | 49.92 | in 289210 / cached 280192 / out 4208 | no | no | no | no | no | no |
| `cli` | `repeat-add` | pass | pass | pass | 10 | 7 | 55.11 | in 358980 / cached 345984 / out 4069 | no | no | no | no | no | no |
| `cli` | `update-existing` | pass | pass | pass | 10 | 5 | 55.08 | in 370545 / cached 359680 / out 4006 | no | no | no | no | no | no |
| `cli` | `bounded-range` | pass | pass | pass | 11 | 7 | 74.02 | in 383489 / cached 356864 / out 6520 | no | no | no | no | no | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 12 | 10 | 106.57 | in 659257 / cached 644736 / out 7983 | no | no | no | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 3 | 5 | 40.01 | in 154781 / cached 148736 / out 4086 | no | no | no | no | yes | no |
| `cli` | `history-limit-two` | pass | pass | pass | 18 | 5 | 64.85 | in 396023 / cached 382208 / out 4851 | no | no | no | no | no | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 5.04 | in 22797 / cached 18816 / out 233 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 1 | 2 | 7.44 | in 46538 / cached 41728 / out 483 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 1 | 2 | 8.60 | in 46630 / cached 41728 / out 603 | no | no | no | no | no | no |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-agentops-production-expanded-r2.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | +0 | -1 | -2.93 | -813 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `repeat-add` | same_pass | +0 | -1 | -1.17 | -387 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `update-existing` | same_pass | +0 | +1 | -6.07 | -2204 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range` | same_pass | +0 | +0 | -0.19 | +99 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range-natural` | same_pass | +0 | +0 | +0.15 | -3189 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | +0 | -1 | -3.43 | -7248 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `history-limit-two` | same_pass | +0 | +0 | +0.49 | +445 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `ambiguous-short-date` | same_pass | +0 | +0 | -2.28 | +512 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | +0 | +0 | +1.35 | -2572 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | same_pass | +1 | +1 | +2.12 | +1554 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `add-two` | same_pass | -5 | -4 | -44.49 | -9301 | same_no | improved_to_no | same_no | same_no | same_no | same_no |
| `cli` | `repeat-add` | same_pass | -2 | +0 | -32.16 | -6448 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `update-existing` | same_pass | -1 | -1 | -7.09 | -18708 | same_no | improved_to_no | same_no | same_no | same_no | same_no |
| `cli` | `bounded-range` | same_pass | +9 | +2 | +43.50 | +23204 | same_no | same_no | same_no | same_no | improved_to_no | same_no |
| `cli` | `bounded-range-natural` | same_pass | +3 | +3 | +57.54 | +1549 | same_no | same_no | same_no | same_no | regressed_to_yes | same_no |
| `cli` | `latest-only` | same_pass | -11 | -1 | -13.17 | -8707 | same_no | same_no | same_no | same_no | regressed_to_yes | same_no |
| `cli` | `history-limit-two` | same_pass | +7 | -1 | +1.77 | +3538 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `ambiguous-short-date` | same_pass | +0 | +0 | -0.05 | +0 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | +0 | +0 | -0.67 | +10 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `non-iso-date-reject` | same_pass | +0 | +0 | +1.16 | +73 | same_no | same_no | same_no | same_no | same_no | same_no |

## Production Stop-Loss

- Triggered: `no`
- Recommendation: `continue_production_hardening`

## Code-First CLI Comparison

- Candidate: `production`
- Baseline: `cli`
- Beats CLI: `yes`
- Recommendation: `prefer_agentops_production_for_routine_weight_operations`

| Criterion | Result | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | pass | 10/10 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine weight scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `total_tools_less_than_or_equal_cli` | pass | production tools 16 vs cli tools 74 |
| `minimum_scenarios_at_or_below_cli` | pass | 10 scenarios at or below CLI tools; required 8 of 10 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 2 | 8 | -6 |
| `repeat-add` | 2 | 10 | -8 |
| `update-existing` | 2 | 10 | -8 |
| `bounded-range` | 2 | 11 | -9 |
| `bounded-range-natural` | 2 | 12 | -10 |
| `latest-only` | 2 | 3 | -1 |
| `history-limit-two` | 2 | 18 | -16 |
| `ambiguous-short-date` | 0 | 0 | +0 |
| `invalid-input` | 1 | 1 | +0 |
| `non-iso-date-reject` | 1 | 1 | +0 |

## Metric Evidence

- `cli/bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.

## Scenario Notes

- `production/add-two`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/add-two/events.jsonl`.
- `production/repeat-add`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/repeat-add/events.jsonl`.
- `production/update-existing`: expected one updated row; observed [2026-03-29 151.6 lb] Raw event reference: `<run-root>/production/update-existing/events.jsonl`.
- `production/bounded-range`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/bounded-range/events.jsonl`.
- `production/bounded-range-natural`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/bounded-range-natural/events.jsonl`.
- `production/latest-only`: expected unchanged seed rows and assistant output limited to latest row 2026-03-30; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/latest-only/events.jsonl`.
- `production/history-limit-two`: expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb, 2026-03-27 154.1 lb] Raw event reference: `<run-root>/production/history-limit-two/events.jsonl`.
- `production/ambiguous-short-date`: expected no write and a year clarification; observed [] Raw event reference: `<run-root>/production/ambiguous-short-date/events.jsonl`.
- `production/invalid-input`: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/production/invalid-input/events.jsonl`.
- `production/non-iso-date-reject`: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/production/non-iso-date-reject/events.jsonl`.
- `cli/add-two`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/cli/add-two/events.jsonl`.
- `cli/repeat-add`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/cli/repeat-add/events.jsonl`.
- `cli/update-existing`: expected one updated row; observed [2026-03-29 151.6 lb] Raw event reference: `<run-root>/cli/update-existing/events.jsonl`.
- `cli/bounded-range`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/cli/bounded-range/events.jsonl`.
- `cli/bounded-range-natural`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/cli/bounded-range-natural/events.jsonl`.
- `cli/latest-only`: expected unchanged seed rows and assistant output limited to latest row 2026-03-30; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/cli/latest-only/events.jsonl`.
- `cli/history-limit-two`: expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb, 2026-03-27 154.1 lb] Raw event reference: `<run-root>/cli/history-limit-two/events.jsonl`.
- `cli/ambiguous-short-date`: expected no write and a year clarification; observed [] Raw event reference: `<run-root>/cli/ambiguous-short-date/events.jsonl`.
- `cli/invalid-input`: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/cli/invalid-input/events.jsonl`.
- `cli/non-iso-date-reject`: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/cli/non-iso-date-reject/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight add/list with a prewarmed per-scenario module cache`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
