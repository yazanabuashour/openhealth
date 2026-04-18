# oh-5yr Agent Eval Results

Date: agentops-bp-final-r3

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-agentops-bp-final-r3.json`

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
| `production` | `add-two` | pass | pass | pass | 3 | 5 | 24.46 | in 141610 / cached 134912 / out 1343 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 3 | 5 | 24.54 | in 140851 / cached 134400 / out 1149 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 3 | 4 | 21.55 | in 140249 / cached 133888 / out 982 | no | no | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 3 | 5 | 22.32 | in 140744 / cached 134400 / out 1041 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 3 | 5 | 21.74 | in 140827 / cached 134400 / out 1100 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 3 | 3 | 23.27 | in 91833 / cached 86016 / out 791 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 3 | 4 | 20.43 | in 92006 / cached 86528 / out 868 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 7.61 | in 44385 / cached 40192 / out 224 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 3 | 5 | 24.50 | in 140656 / cached 134400 / out 1140 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 3 | 5 | 24.30 | in 141117 / cached 129792 / out 1173 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 3 | 4 | 26.66 | in 108857 / cached 94720 / out 1321 | no | no | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 3 | 4 | 20.13 | in 92050 / cached 86528 / out 778 | no | no | no | no | no | no |
| `production` | `bp-history-limit-two` | pass | pass | pass | 3 | 6 | 23.04 | in 140617 / cached 134400 / out 963 | no | no | no | no | no | no |
| `production` | `bp-bounded-range` | pass | pass | pass | 3 | 4 | 21.05 | in 92310 / cached 86528 / out 889 | no | no | no | no | no | no |
| `production` | `bp-bounded-range-natural` | pass | pass | pass | 3 | 4 | 28.76 | in 92233 / cached 86528 / out 937 | no | no | no | no | no | no |
| `production` | `bp-invalid-input` | pass | pass | pass | 3 | 5 | 22.48 | in 140996 / cached 134400 / out 1117 | no | no | no | no | no | no |
| `production` | `bp-non-iso-date-reject` | pass | pass | pass | 3 | 5 | 23.32 | in 141867 / cached 135424 / out 1350 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 4 | 6 | 30.07 | in 136736 / cached 130816 / out 879 | no | no | no | no | yes | no |
| `cli` | `repeat-add` | pass | pass | pass | 4 | 4 | 23.46 | in 158821 / cached 153216 / out 758 | no | no | no | no | yes | no |
| `cli` | `update-existing` | pass | pass | pass | 7 | 5 | 25.00 | in 131308 / cached 118656 / out 1132 | no | yes | no | no | yes | no |
| `cli` | `bounded-range` | pass | pass | pass | 2 | 5 | 16.84 | in 115007 / cached 109440 / out 591 | no | no | no | no | yes | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 3 | 17.10 | in 67175 / cached 62592 / out 476 | no | no | no | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 2 | 5 | 16.35 | in 112318 / cached 107392 / out 523 | no | no | no | no | yes | no |
| `cli` | `history-limit-two` | pass | pass | pass | 2 | 4 | 21.24 | in 112502 / cached 107392 / out 572 | no | no | no | no | yes | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 6.29 | in 44157 / cached 40192 / out 316 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 1 | 2 | 9.62 | in 44064 / cached 40192 / out 551 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 1 | 2 | 9.88 | in 44116 / cached 40192 / out 573 | no | no | no | no | no | no |
| `cli` | `bp-add-two` | fail | fail | pass | 5 | 5 | 30.77 | in 182958 / cached 176640 / out 1235 | no | no | no | no | yes | no |
| `cli` | `bp-latest-only` | pass | pass | pass | 2 | 3 | 18.58 | in 66737 / cached 60032 / out 355 | no | no | no | no | yes | no |
| `cli` | `bp-history-limit-two` | pass | pass | pass | 2 | 3 | 16.26 | in 66766 / cached 62592 / out 356 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range` | pass | pass | pass | 2 | 4 | 18.72 | in 115035 / cached 108928 / out 521 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range-natural` | pass | pass | pass | 2 | 4 | 17.75 | in 112409 / cached 107392 / out 489 | no | no | no | no | yes | no |
| `cli` | `bp-invalid-input` | pass | pass | pass | 1 | 2 | 9.64 | in 44068 / cached 40192 / out 355 | no | no | no | no | no | no |
| `cli` | `bp-non-iso-date-reject` | pass | pass | pass | 2 | 4 | 19.75 | in 112895 / cached 107904 / out 699 | no | no | no | no | yes | no |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-agentops-bp-final-r2.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | +0 | +1 | -8.77 | +845 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `repeat-add` | same_pass | +0 | +0 | -6.98 | +125 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `update-existing` | same_pass | -1 | -1 | -3.95 | -594 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range` | same_pass | +0 | -1 | +2.01 | +12 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range-natural` | same_pass | +0 | +0 | -3.37 | -150 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | +1 | -1 | +2.76 | +593 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `history-limit-two` | same_pass | +0 | -1 | -2.56 | -323 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `ambiguous-short-date` | same_pass | +0 | -1 | -6.24 | -27284 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | +0 | +0 | +0.17 | +377 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | same_pass | +0 | +0 | -1.22 | +4846 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | same_pass | +0 | -1 | -4.55 | +7859 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-latest-only` | same_pass | +0 | +0 | -3.50 | -627 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-history-limit-two` | same_pass | +0 | +2 | +2.08 | +596 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range` | same_pass | +0 | -1 | -5.44 | -1201 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range-natural` | same_pass | +0 | -2 | +6.24 | -692 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-invalid-input` | same_pass | +0 | -1 | -4.67 | +120 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-non-iso-date-reject` | same_pass | +0 | +1 | -3.84 | +547 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `add-two` | same_pass | +0 | +1 | +7.97 | +222 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `repeat-add` | same_pass | +0 | +0 | +1.03 | -345 | same_no | improved_to_no | same_no | same_no | same_yes | same_no |
| `cli` | `update-existing` | same_pass | -2 | -1 | -3.11 | -2817 | same_no | same_yes | improved_to_no | same_no | same_yes | same_no |
| `cli` | `bounded-range` | same_pass | +0 | +0 | -2.10 | -457 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bounded-range-natural` | same_pass | +0 | -2 | -2.04 | -422 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `latest-only` | same_pass | +0 | +2 | -0.53 | +811 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `history-limit-two` | same_pass | +0 | +1 | +2.77 | +981 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `ambiguous-short-date` | same_pass | +0 | +0 | -0.78 | +4 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | +0 | +0 | +0.32 | -14 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `non-iso-date-reject` | same_pass | -1 | -2 | -16.18 | -950 | same_no | same_no | same_no | same_no | improved_to_no | same_no |
| `cli` | `bp-add-two` | regressed | +3 | +1 | +10.83 | +824 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-latest-only` | same_pass | +0 | +0 | -0.68 | +2570 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-history-limit-two` | same_pass | +0 | +0 | -2.35 | +19 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range` | same_pass | +0 | -1 | +2.05 | +854 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range-natural` | same_pass | +0 | +0 | -3.62 | +384 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-invalid-input` | same_pass | +0 | +0 | +1.06 | -10 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `bp-non-iso-date-reject` | same_pass | +0 | +0 | -1.88 | -123 | same_no | same_no | same_no | same_no | same_yes | same_no |

## Production Stop-Loss

- Triggered: `yes`
- Recommendation: `continue_cli_baseline_for_agent_operations`
- Trigger: production used more than 2x CLI tools in bp-invalid-input, invalid-input, non-iso-date-reject

## Code-First CLI Comparison

- Candidate: `production`
- Baseline: `cli`
- Beats CLI: `no`
- Recommendation: `continue_cli_for_routine_openhealth_operations`

| Criterion | Result | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | pass | 17/17 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `total_tools_less_than_or_equal_cli` | fail | production tools 49 vs cli tools 42 |
| `minimum_scenarios_at_or_below_cli` | fail | 5 scenarios at or below CLI tools; required 14 of 17 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 3 | 4 | -1 |
| `repeat-add` | 3 | 4 | -1 |
| `update-existing` | 3 | 7 | -4 |
| `bounded-range` | 3 | 2 | +1 |
| `bounded-range-natural` | 3 | 2 | +1 |
| `latest-only` | 3 | 2 | +1 |
| `history-limit-two` | 3 | 2 | +1 |
| `ambiguous-short-date` | 1 | 1 | +0 |
| `invalid-input` | 3 | 1 | +2 |
| `non-iso-date-reject` | 3 | 1 | +2 |
| `bp-add-two` | 3 | 5 | -2 |
| `bp-latest-only` | 3 | 2 | +1 |
| `bp-history-limit-two` | 3 | 2 | +1 |
| `bp-bounded-range` | 3 | 2 | +1 |
| `bp-bounded-range-natural` | 3 | 2 | +1 |
| `bp-invalid-input` | 3 | 1 | +2 |
| `bp-non-iso-date-reject` | 3 | 2 | +1 |

## Metric Evidence

- `cli/add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/repeat-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/update-existing` broad repo search: `/bin/zsh -lc "rg -n \"weight (add|list)|OPENHEALTH_DATABASE_PATH|sqlite\" cmd . -g '"'!**/node_modules/**'"'"`.
- `cli/update-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 151.6 --unit lb && go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 2'`.
- `cli/bp-add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 118 --diastolic 76'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/bp-history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 2'`.
- `cli/bp-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-non-iso-date-reject` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026/03/31 --systolic 122 --diastolic 78'`.

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
- `production/bp-add-two`: expected exactly two newest-first blood-pressure rows; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/bp-add-two/events.jsonl`.
- `production/bp-latest-only`: expected unchanged seed rows and assistant output limited to latest blood-pressure row 2026-03-30; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/bp-latest-only/events.jsonl`.
- `production/bp-history-limit-two`: expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80, 2026-03-27 126/82] Raw event reference: `<run-root>/production/bp-history-limit-two/events.jsonl`.
- `production/bp-bounded-range`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/bp-bounded-range/events.jsonl`.
- `production/bp-bounded-range-natural`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/bp-bounded-range-natural/events.jsonl`.
- `production/bp-invalid-input`: expected no write and an invalid blood-pressure rejection; observed [] Raw event reference: `<run-root>/production/bp-invalid-input/events.jsonl`.
- `production/bp-non-iso-date-reject`: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/production/bp-non-iso-date-reject/events.jsonl`.
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
- `cli/bp-add-two`: expected exactly two newest-first blood-pressure rows; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/bp-add-two/events.jsonl`.
- `cli/bp-latest-only`: expected unchanged seed rows and assistant output limited to latest blood-pressure row 2026-03-30; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/bp-latest-only/events.jsonl`.
- `cli/bp-history-limit-two`: expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80, 2026-03-27 126/82] Raw event reference: `<run-root>/cli/bp-history-limit-two/events.jsonl`.
- `cli/bp-bounded-range`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/bp-bounded-range/events.jsonl`.
- `cli/bp-bounded-range-natural`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/cli/bp-bounded-range-natural/events.jsonl`.
- `cli/bp-invalid-input`: expected no write and an invalid blood-pressure rejection; observed [] Raw event reference: `<run-root>/cli/bp-invalid-input/events.jsonl`.
- `cli/bp-non-iso-date-reject`: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/cli/bp-non-iso-date-reject/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands with a prewarmed per-scenario module cache`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
