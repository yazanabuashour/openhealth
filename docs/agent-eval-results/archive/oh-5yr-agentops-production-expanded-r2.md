# oh-5yr Agent Eval Results

Date: agentops-production-expanded-r2

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/archive/oh-5yr-agentops-production-expanded-r2.json`

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
| `production` | `add-two` | pass | pass | pass | 2 | 4 | 23.70 | in 123270 / cached 116608 / out 1134 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 2 | 4 | 21.70 | in 122665 / cached 118656 / out 979 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 2 | 3 | 25.13 | in 72818 / cached 66688 / out 1301 | no | no | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 2 | 4 | 20.01 | in 122985 / cached 116608 / out 1049 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 2 | 4 | 21.68 | in 122353 / cached 112512 / out 856 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 2 | 4 | 22.49 | in 122569 / cached 112000 / out 960 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 2 | 4 | 19.62 | in 122843 / cached 116608 / out 1002 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 6.60 | in 22816 / cached 19328 / out 270 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 1 | 2 | 6.90 | in 47199 / cached 41728 / out 432 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 0 | 1 | 7.79 | in 22832 / cached 18816 / out 353 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 13 | 10 | 94.41 | in 559631 / cached 541312 / out 7516 | no | yes | no | no | no | no |
| `cli` | `repeat-add` | pass | pass | pass | 12 | 7 | 87.27 | in 527732 / cached 508288 / out 7562 | no | no | no | no | no | no |
| `cli` | `update-existing` | pass | pass | pass | 11 | 6 | 62.17 | in 510853 / cached 481280 / out 5490 | no | yes | no | no | no | no |
| `cli` | `bounded-range` | pass | pass | pass | 2 | 5 | 30.52 | in 125149 / cached 121728 / out 2666 | no | no | no | no | yes | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 9 | 7 | 49.03 | in 247852 / cached 234880 / out 4883 | no | no | no | no | no | no |
| `cli` | `latest-only` | pass | pass | pass | 14 | 6 | 53.18 | in 442400 / cached 427648 / out 3419 | no | no | no | no | no | no |
| `cli` | `history-limit-two` | pass | pass | pass | 11 | 6 | 63.08 | in 334885 / cached 324608 / out 5193 | no | no | no | no | no | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 5.09 | in 22797 / cached 18816 / out 294 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 1 | 2 | 8.11 | in 46528 / cached 41728 / out 357 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 1 | 2 | 7.44 | in 46557 / cached 41728 / out 426 | no | no | no | no | no | no |

## Comparison

Baseline: `docs/agent-eval-results/archive/oh-5yr-agentops-production-expanded-r1.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | -1 | +0 | -4.77 | -1508 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `repeat-add` | same_pass | +0 | +0 | +3.48 | -2280 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `update-existing` | same_pass | +0 | -1 | +2.44 | -6700 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range` | same_pass | +0 | +0 | -1.19 | -95 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range-natural` | same_pass | +0 | +1 | +0.36 | +6258 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | +0 | +0 | +3.69 | +4357 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `history-limit-two` | same_pass | +0 | +1 | -2.38 | -1394 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `ambiguous-short-date` | same_pass | +0 | +0 | +2.07 | -512 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | +0 | +0 | -4.14 | +372 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | same_pass | -1 | -1 | -14.86 | -1508 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `add-two` | same_pass | +2 | +1 | -12.75 | +2716 | same_no | same_yes | improved_to_no | same_no | same_no | same_no |
| `cli` | `repeat-add` | same_pass | -3 | -8 | -6.86 | -11193 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `update-existing` | same_pass | +2 | +1 | +9.63 | -286 | same_no | regressed_to_yes | same_no | same_no | improved_to_no | same_no |
| `cli` | `bounded-range` | same_pass | -9 | -5 | -47.16 | -23892 | same_no | same_no | same_no | same_no | regressed_to_yes | same_no |
| `cli` | `bounded-range-natural` | same_pass | -12 | -3 | -61.51 | -1024 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `latest-only` | same_pass | +1 | -8 | -46.70 | -10378 | same_no | improved_to_no | same_no | same_no | same_no | same_no |
| `cli` | `history-limit-two` | same_pass | +0 | -4 | -10.63 | -6606 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `ambiguous-short-date` | same_pass | +0 | +0 | -0.83 | +2560 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | +0 | -1 | -5.41 | +2456 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `non-iso-date-reject` | same_pass | +1 | +1 | +0.83 | +3392 | same_no | same_no | same_no | same_no | same_no | same_no |

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
| `total_tools_less_than_or_equal_cli` | pass | production tools 15 vs cli tools 74 |
| `minimum_scenarios_at_or_below_cli` | pass | 10 scenarios at or below CLI tools; required 8 of 10 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 2 | 13 | -11 |
| `repeat-add` | 2 | 12 | -10 |
| `update-existing` | 2 | 11 | -9 |
| `bounded-range` | 2 | 2 | +0 |
| `bounded-range-natural` | 2 | 9 | -7 |
| `latest-only` | 2 | 14 | -12 |
| `history-limit-two` | 2 | 11 | -9 |
| `ambiguous-short-date` | 0 | 0 | +0 |
| `invalid-input` | 1 | 1 | +0 |
| `non-iso-date-reject` | 0 | 1 | -1 |

## Metric Evidence

- `cli/add-two` broad repo search: `/bin/zsh -lc 'rg -n "unsupported weight task action|WeightTaskRequest|WeightListMode|RunWeightTask" .'`.
- `cli/update-existing` broad repo search: `/bin/zsh -lc "go doc -all ./agentops 2>/dev/null | rg -n -A20 -B4 'WeightTaskAction'"`.
- `cli/bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.

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
