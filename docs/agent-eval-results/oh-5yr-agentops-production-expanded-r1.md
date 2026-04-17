# oh-5yr Agent Eval Results

Date: agentops-production-expanded-r1

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-agentops-production-expanded-r1.json`

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
| `production` | `add-two` | pass | pass | pass | 3 | 4 | 28.47 | in 178282 / cached 170112 / out 2090 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 2 | 4 | 18.22 | in 97425 / cached 91136 / out 870 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 2 | 4 | 22.69 | in 123294 / cached 110464 / out 1183 | no | no | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 2 | 4 | 21.20 | in 122568 / cached 116096 / out 885 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 2 | 3 | 21.32 | in 72319 / cached 68736 / out 858 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 2 | 4 | 18.80 | in 97348 / cached 91136 / out 883 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 2 | 3 | 22.00 | in 72269 / cached 64640 / out 834 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 4.53 | in 22816 / cached 18816 / out 257 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 1 | 2 | 11.04 | in 47339 / cached 42240 / out 604 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 1 | 2 | 22.65 | in 47252 / cached 41728 / out 372 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 11 | 9 | 107.16 | in 621683 / cached 606080 / out 8684 | no | yes | yes | no | no | no |
| `cli` | `repeat-add` | pass | pass | pass | 15 | 15 | 94.13 | in 710317 / cached 679680 / out 9283 | no | no | no | no | no | no |
| `cli` | `update-existing` | pass | pass | pass | 9 | 5 | 52.54 | in 268707 / cached 238848 / out 4343 | no | no | no | no | yes | no |
| `cli` | `bounded-range` | pass | pass | pass | 11 | 10 | 77.68 | in 710321 / cached 683008 / out 6072 | no | no | no | no | no | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 21 | 10 | 110.54 | in 682284 / cached 668288 / out 10475 | no | no | no | no | no | no |
| `cli` | `latest-only` | pass | pass | pass | 13 | 14 | 99.88 | in 685226 / cached 660096 / out 8975 | no | yes | no | no | no | no |
| `cli` | `history-limit-two` | pass | pass | pass | 11 | 10 | 73.71 | in 524019 / cached 507136 / out 6796 | no | no | no | no | no | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 5.92 | in 22797 / cached 21376 / out 307 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 1 | 3 | 13.52 | in 46632 / cached 44288 / out 1357 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 0 | 1 | 6.61 | in 22813 / cached 21376 / out 490 | no | no | no | no | no | no |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-code-first-iter3.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `repeat-add` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `update-existing` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bounded-range-natural` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `latest-only` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `history-limit-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `ambiguous-short-date` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `invalid-input` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `non-iso-date-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `add-two` | same_pass | +4 | +4 | +75.76 | -8994 | same_no | same_yes | regressed_to_yes | same_no | same_no | same_no |
| `cli` | `repeat-add` | same_pass | +9 | +8 | +50.94 | +18209 | same_no | improved_to_no | same_no | same_no | same_no | same_no |
| `cli` | `update-existing` | same_pass | +1 | -1 | +9.35 | +11728 | same_no | improved_to_no | improved_to_no | same_no | regressed_to_yes | same_no |
| `cli` | `bounded-range` | same_pass | +9 | +6 | +53.39 | +21727 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `bounded-range-natural` | same_pass | +19 | +6 | +90.82 | +8227 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `latest-only` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `history-limit-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `ambiguous-short-date` | same_pass | -1 | -1 | -1.74 | -3061 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | -1 | +1 | +4.54 | -2191 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `non-iso-date-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |

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
| `total_tools_less_than_or_equal_cli` | pass | production tools 17 vs cli tools 92 |
| `minimum_scenarios_at_or_below_cli` | pass | 9 scenarios at or below CLI tools; required 8 of 10 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 3 | 11 | -8 |
| `repeat-add` | 2 | 15 | -13 |
| `update-existing` | 2 | 9 | -7 |
| `bounded-range` | 2 | 11 | -9 |
| `bounded-range-natural` | 2 | 21 | -19 |
| `latest-only` | 2 | 13 | -11 |
| `history-limit-two` | 2 | 11 | -9 |
| `ambiguous-short-date` | 0 | 0 | +0 |
| `invalid-input` | 1 | 1 | +0 |
| `non-iso-date-reject` | 1 | 0 | +1 |

## Metric Evidence

- `cli/add-two` broad repo search: `/bin/zsh -lc 'rg -n "WeightTask|WeightListMode|action" .'`.
- `cli/add-two` generated path from broad search: `/bin/zsh -lc 'rg -n "WeightTask|WeightListMode|action" .'`.
- `cli/update-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-29'`.
- `cli/latest-only` broad repo search: `/bin/zsh -lc 'rg -n "unsupported weight task action|WeightTaskRequest|WeightListModeLatest|WeightListModeHistory" .'`.

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
