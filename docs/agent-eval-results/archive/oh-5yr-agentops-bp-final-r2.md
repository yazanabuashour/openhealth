# oh-5yr Agent Eval Results

Date: agentops-bp-final-r2

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-agentops-bp-final-r2.json`

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
| `production` | `add-two` | pass | pass | pass | 3 | 4 | 33.23 | in 93405 / cached 87552 / out 1778 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 3 | 5 | 31.52 | in 140726 / cached 134400 / out 1096 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 4 | 5 | 25.50 | in 166827 / cached 159872 / out 1744 | no | no | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 3 | 6 | 20.31 | in 140732 / cached 134400 / out 1028 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 3 | 5 | 25.11 | in 140977 / cached 134400 / out 1138 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 2 | 4 | 20.51 | in 114152 / cached 108928 / out 749 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 3 | 5 | 22.99 | in 116265 / cached 110464 / out 995 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 1 | 3 | 13.85 | in 52725 / cached 21248 / out 480 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 3 | 5 | 24.33 | in 116343 / cached 110464 / out 1127 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 3 | 5 | 25.52 | in 140879 / cached 134400 / out 1115 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 3 | 5 | 31.21 | in 116742 / cached 110464 / out 1112 | no | no | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 3 | 4 | 23.63 | in 141061 / cached 134912 / out 1039 | no | no | no | no | no | no |
| `production` | `bp-history-limit-two` | pass | pass | pass | 3 | 4 | 20.96 | in 92149 / cached 86528 / out 864 | no | no | no | no | no | no |
| `production` | `bp-bounded-range` | pass | pass | pass | 3 | 5 | 26.49 | in 141383 / cached 134400 / out 1139 | no | no | no | no | no | no |
| `production` | `bp-bounded-range-natural` | pass | pass | pass | 3 | 6 | 22.52 | in 140797 / cached 134400 / out 1078 | no | no | no | no | no | no |
| `production` | `bp-invalid-input` | pass | pass | pass | 3 | 6 | 27.15 | in 140876 / cached 134400 / out 1149 | no | no | no | no | no | no |
| `production` | `bp-non-iso-date-reject` | pass | pass | pass | 3 | 4 | 27.16 | in 92424 / cached 86528 / out 1053 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 4 | 5 | 22.10 | in 158914 / cached 153216 / out 766 | no | no | no | no | yes | no |
| `cli` | `repeat-add` | pass | pass | pass | 4 | 4 | 22.43 | in 161214 / cached 155264 / out 1171 | no | yes | no | no | yes | no |
| `cli` | `update-existing` | pass | pass | pass | 9 | 6 | 28.11 | in 233965 / cached 218496 / out 1413 | no | yes | yes | no | yes | no |
| `cli` | `bounded-range` | pass | pass | pass | 2 | 5 | 18.94 | in 114952 / cached 108928 / out 530 | no | no | no | no | yes | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 5 | 19.14 | in 112397 / cached 107392 / out 544 | no | no | no | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 2 | 3 | 16.88 | in 66707 / cached 62592 / out 394 | no | no | no | no | yes | no |
| `cli` | `history-limit-two` | pass | pass | pass | 2 | 3 | 18.47 | in 66721 / cached 62592 / out 388 | no | no | no | no | yes | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 7.07 | in 44153 / cached 40192 / out 347 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 1 | 2 | 9.30 | in 44078 / cached 40192 / out 461 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 2 | 4 | 26.06 | in 89866 / cached 84992 / out 718 | no | no | no | no | yes | no |
| `cli` | `bp-add-two` | pass | pass | pass | 2 | 4 | 19.94 | in 112886 / cached 107392 / out 662 | no | no | no | no | yes | no |
| `cli` | `bp-latest-only` | pass | pass | pass | 2 | 3 | 19.26 | in 66727 / cached 62592 / out 350 | no | no | no | no | yes | no |
| `cli` | `bp-history-limit-two` | pass | pass | pass | 2 | 3 | 18.61 | in 66747 / cached 62592 / out 364 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range` | pass | pass | pass | 2 | 5 | 16.67 | in 112645 / cached 107392 / out 484 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range-natural` | pass | pass | pass | 2 | 4 | 21.37 | in 89625 / cached 84992 / out 515 | no | no | no | no | yes | no |
| `cli` | `bp-invalid-input` | pass | pass | pass | 1 | 2 | 8.58 | in 44078 / cached 40192 / out 398 | no | no | no | no | no | no |
| `cli` | `bp-non-iso-date-reject` | pass | pass | pass | 2 | 4 | 21.63 | in 113530 / cached 108416 / out 917 | no | no | no | no | yes | no |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-agentops-bp-final-r1.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | +0 | +0 | +12.79 | -717 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `repeat-add` | same_pass | +0 | +1 | +9.92 | -5315 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `update-existing` | same_pass | +1 | +1 | -0.27 | +1452 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range` | same_pass | +0 | +1 | -5.00 | -4928 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range-natural` | same_pass | +0 | +1 | -0.50 | -83 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | +0 | +0 | -6.04 | -56 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `history-limit-two` | same_pass | +0 | +1 | -1.76 | -535 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `ambiguous-short-date` | same_pass | +0 | +1 | +8.22 | +27724 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | +0 | +0 | -1.90 | -347 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | same_pass | +0 | +0 | -0.93 | +321 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | same_pass | +0 | +0 | +8.75 | -5736 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-latest-only` | same_pass | +0 | -1 | +1.00 | -5032 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-history-limit-two` | same_pass | +0 | -2 | -1.41 | -719 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range` | same_pass | +0 | +0 | +2.99 | +215 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range-natural` | same_pass | +0 | +2 | -0.35 | -256 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-invalid-input` | same_pass | +0 | +1 | -0.60 | -3330 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-non-iso-date-reject` | same_pass | -1 | -2 | +0.35 | -1659 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `add-two` | same_pass | +2 | +1 | -1.34 | +298 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `repeat-add` | same_pass | +2 | +0 | +3.94 | +557 | same_no | regressed_to_yes | same_no | same_no | same_yes | same_no |
| `cli` | `update-existing` | same_pass | +1 | +0 | +1.23 | +5642 | same_no | same_yes | same_yes | same_no | same_yes | same_no |
| `cli` | `bounded-range` | same_pass | +0 | +0 | -0.64 | +368 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bounded-range-natural` | same_pass | +0 | +1 | +1.14 | -43 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `latest-only` | same_pass | +0 | +0 | +0.44 | +76 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `history-limit-two` | same_pass | +0 | -1 | -0.16 | -798 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `ambiguous-short-date` | same_pass | +0 | +0 | +1.89 | +25 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | +0 | +0 | +2.78 | +4 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `non-iso-date-reject` | same_pass | +1 | +2 | +17.01 | +901 | same_no | same_no | same_no | same_no | regressed_to_yes | same_no |
| `cli` | `bp-add-two` | same_pass | +0 | -1 | -4.27 | -320 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-latest-only` | same_pass | +0 | -1 | -1.58 | -307 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-history-limit-two` | same_pass | +0 | -1 | -5.79 | -365 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range` | same_pass | +0 | +1 | -1.79 | -845 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range-natural` | same_pass | -2 | -3 | -5.06 | -3380 | same_no | improved_to_no | improved_to_no | same_no | same_yes | same_no |
| `cli` | `bp-invalid-input` | same_pass | -2 | -3 | -15.62 | -4675 | same_no | improved_to_no | improved_to_no | same_no | improved_to_no | same_no |
| `cli` | `bp-non-iso-date-reject` | same_pass | +0 | +1 | +1.89 | +800 | same_no | same_no | same_no | same_no | same_yes | same_no |

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
| `update-existing` | 4 | 9 | -5 |
| `bounded-range` | 3 | 2 | +1 |
| `bounded-range-natural` | 3 | 2 | +1 |
| `latest-only` | 2 | 2 | +0 |
| `history-limit-two` | 3 | 2 | +1 |
| `ambiguous-short-date` | 1 | 1 | +0 |
| `invalid-input` | 3 | 1 | +2 |
| `non-iso-date-reject` | 3 | 2 | +1 |
| `bp-add-two` | 3 | 2 | +1 |
| `bp-latest-only` | 3 | 2 | +1 |
| `bp-history-limit-two` | 3 | 2 | +1 |
| `bp-bounded-range` | 3 | 2 | +1 |
| `bp-bounded-range-natural` | 3 | 2 | +1 |
| `bp-invalid-input` | 3 | 1 | +2 |
| `bp-non-iso-date-reject` | 3 | 2 | +1 |

## Metric Evidence

- `cli/add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/repeat-add` broad repo search: `/bin/zsh -lc 'pwd && rg -n "" .agents/skills/openhealth/SKILL.md'`.
- `cli/repeat-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/update-existing` broad repo search: `/bin/zsh -lc 'rg -n "weight add|Replace|upsert|duplicate|same date|ON CONFLICT|weight" cmd internal .'`.
- `cli/update-existing` generated path from broad search: `/bin/zsh -lc 'rg -n "weight add|Replace|upsert|duplicate|same date|ON CONFLICT|weight" cmd internal .'`.
- `cli/update-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight --help'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-29'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 151.6 --unit lb'`.
- `cli/bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 2'`.
- `cli/non-iso-date-reject` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026/03/31 --value 152.2 --unit lb'`.
- `cli/bp-add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64 && go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 118 --diastolic 76 && go run ./cmd/openhealth blood-pressure list --limit 25'`.
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
- `cli/bp-add-two`: expected exactly two newest-first blood-pressure rows; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/cli/bp-add-two/events.jsonl`.
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
