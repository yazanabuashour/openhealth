# oh-5yr Agent Eval Results

Date: agentops-bp-final-r1

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/archive/oh-5yr-agentops-bp-final-r1.json`

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
| `production` | `add-two` | pass | pass | pass | 3 | 4 | 20.44 | in 140970 / cached 134400 / out 1120 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 3 | 4 | 21.60 | in 140921 / cached 129280 / out 1120 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 3 | 4 | 25.77 | in 92031 / cached 86528 / out 912 | no | no | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 3 | 5 | 25.31 | in 116604 / cached 105344 / out 935 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 3 | 4 | 25.61 | in 141060 / cached 134400 / out 1246 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 2 | 4 | 26.55 | in 114208 / cached 108928 / out 792 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 3 | 4 | 24.75 | in 140224 / cached 133888 / out 954 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 5.63 | in 44457 / cached 40704 / out 299 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 3 | 5 | 26.23 | in 140626 / cached 134400 / out 1226 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 3 | 5 | 26.45 | in 142094 / cached 135936 / out 1389 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 3 | 5 | 22.46 | in 141294 / cached 129280 / out 1105 | no | no | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 3 | 5 | 22.63 | in 140461 / cached 129280 / out 900 | no | no | no | no | no | no |
| `production` | `bp-history-limit-two` | pass | pass | pass | 3 | 6 | 22.37 | in 140740 / cached 134400 / out 1016 | no | no | no | no | no | no |
| `production` | `bp-bounded-range` | pass | pass | pass | 3 | 5 | 23.50 | in 141168 / cached 134400 / out 1035 | no | no | no | no | no | no |
| `production` | `bp-bounded-range-natural` | pass | pass | pass | 3 | 4 | 22.87 | in 141053 / cached 134400 / out 1096 | no | no | no | no | no | no |
| `production` | `bp-invalid-input` | pass | pass | pass | 3 | 5 | 27.75 | in 116686 / cached 106880 / out 1089 | no | no | no | no | no | no |
| `production` | `bp-non-iso-date-reject` | pass | pass | pass | 4 | 6 | 26.81 | in 164355 / cached 156800 / out 1651 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 2 | 4 | 23.44 | in 112792 / cached 107392 / out 633 | no | no | no | no | yes | no |
| `cli` | `repeat-add` | pass | pass | pass | 2 | 4 | 18.49 | in 112785 / cached 107392 / out 651 | no | no | no | no | yes | no |
| `cli` | `update-existing` | pass | pass | pass | 8 | 6 | 26.88 | in 208995 / cached 199168 / out 1347 | no | yes | yes | no | yes | no |
| `cli` | `bounded-range` | pass | pass | pass | 2 | 5 | 19.58 | in 115096 / cached 109440 / out 576 | no | no | no | no | yes | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 4 | 18.00 | in 112440 / cached 107392 / out 510 | no | no | no | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 2 | 3 | 16.44 | in 66631 / cached 62592 / out 277 | no | no | no | no | yes | no |
| `cli` | `history-limit-two` | pass | pass | pass | 2 | 4 | 18.63 | in 112319 / cached 107392 / out 485 | no | no | no | no | yes | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 5.18 | in 44128 / cached 40192 / out 270 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 1 | 2 | 6.52 | in 44074 / cached 40192 / out 466 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 1 | 2 | 9.05 | in 44165 / cached 40192 / out 669 | no | no | no | no | no | no |
| `cli` | `bp-add-two` | pass | pass | pass | 2 | 5 | 24.21 | in 115766 / cached 109952 / out 844 | no | no | no | no | yes | no |
| `cli` | `bp-latest-only` | pass | pass | pass | 2 | 4 | 20.84 | in 89434 / cached 84992 / out 414 | no | no | no | no | yes | no |
| `cli` | `bp-history-limit-two` | pass | pass | pass | 2 | 4 | 24.40 | in 89512 / cached 84992 / out 526 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range` | pass | pass | pass | 2 | 4 | 18.46 | in 115026 / cached 108928 / out 503 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range-natural` | pass | pass | pass | 4 | 7 | 26.43 | in 170957 / cached 162944 / out 990 | no | yes | yes | no | yes | no |
| `cli` | `bp-invalid-input` | pass | pass | pass | 3 | 5 | 24.20 | in 123633 / cached 115072 / out 923 | no | yes | yes | no | yes | no |
| `cli` | `bp-non-iso-date-reject` | pass | pass | pass | 2 | 3 | 19.74 | in 66906 / cached 62592 / out 616 | no | no | no | no | yes | no |

## Comparison

Baseline: `docs/agent-eval-results/archive/oh-5yr-agentops-bp-baseline-r1.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | +0 | +0 | -2.66 | +215 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `repeat-add` | same_pass | -1 | -1 | -1.47 | +4725 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `update-existing` | same_pass | -5 | -2 | -11.33 | -7878 | same_no | improved_to_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range` | same_pass | +0 | +2 | +4.59 | +5601 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range-natural` | same_pass | +0 | +0 | +5.38 | +1102 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | +0 | +1 | +6.16 | -18079 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `history-limit-two` | same_pass | -1 | +1 | +2.14 | +762 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `ambiguous-short-date` | same_pass | +0 | +0 | -0.74 | +7 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | +0 | +1 | -1.20 | +952 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | same_pass | +0 | +2 | +4.18 | +483 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | same_pass | -3 | -1 | -15.24 | -183 | same_no | improved_to_no | same_no | same_no | same_no | same_no |
| `production` | `bp-latest-only` | same_pass | +0 | +1 | +1.46 | +5586 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-history-limit-two` | same_pass | +0 | +1 | -3.81 | +210 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range` | same_pass | +0 | +0 | -0.04 | +19 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range-natural` | fixed | +0 | +0 | -3.21 | +870 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-invalid-input` | same_pass | +0 | +0 | +4.95 | -10990 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-non-iso-date-reject` | same_pass | +1 | +1 | +5.29 | +1440 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `add-two` | same_pass | -4 | -3 | -10.49 | -4185 | same_no | improved_to_no | improved_to_no | same_no | same_yes | same_no |
| `cli` | `repeat-add` | same_pass | -8 | -3 | -15.72 | -9413 | same_no | improved_to_no | improved_to_no | same_no | same_yes | same_no |
| `cli` | `update-existing` | same_pass | -6 | -2 | -11.05 | -7103 | same_no | same_yes | same_yes | same_no | same_yes | same_no |
| `cli` | `bounded-range` | same_pass | -6 | -1 | -12.96 | -20722 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bounded-range-natural` | same_pass | -7 | -2 | -16.43 | -8196 | same_no | improved_to_no | improved_to_no | same_no | same_yes | same_no |
| `cli` | `latest-only` | same_pass | -2 | -1 | -1.00 | -3508 | same_no | improved_to_no | same_no | same_no | same_yes | same_no |
| `cli` | `history-limit-two` | same_pass | -5 | -2 | -9.60 | -6473 | same_no | improved_to_no | improved_to_no | same_no | same_yes | same_no |
| `cli` | `ambiguous-short-date` | same_pass | +0 | +0 | -1.69 | +35 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | +0 | +0 | +0.60 | +52 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `non-iso-date-reject` | same_pass | -1 | -2 | -11.06 | -597 | same_no | same_no | same_no | same_no | improved_to_no | same_no |
| `cli` | `bp-add-two` | same_pass | -2 | +1 | -1.30 | +457 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-latest-only` | same_pass | -5 | -1 | -6.53 | -11352 | same_no | improved_to_no | improved_to_no | same_no | same_yes | same_no |
| `cli` | `bp-history-limit-two` | same_pass | -6 | -1 | -1.91 | -5746 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range` | same_pass | -4 | -1 | -5.89 | -4629 | same_no | improved_to_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range-natural` | same_pass | -12 | -2 | -59.69 | -11439 | same_no | same_yes | same_yes | same_no | same_yes | same_no |
| `cli` | `bp-invalid-input` | same_pass | +1 | +1 | +1.51 | +3555 | same_no | regressed_to_yes | regressed_to_yes | same_no | same_yes | same_no |
| `cli` | `bp-non-iso-date-reject` | same_pass | -1 | -1 | -3.31 | -5336 | same_no | same_no | same_no | same_no | same_yes | same_no |

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
| `total_tools_less_than_or_equal_cli` | fail | production tools 49 vs cli tools 40 |
| `minimum_scenarios_at_or_below_cli` | fail | 5 scenarios at or below CLI tools; required 14 of 17 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 3 | 2 | +1 |
| `repeat-add` | 3 | 2 | +1 |
| `update-existing` | 3 | 8 | -5 |
| `bounded-range` | 3 | 2 | +1 |
| `bounded-range-natural` | 3 | 2 | +1 |
| `latest-only` | 2 | 2 | +0 |
| `history-limit-two` | 3 | 2 | +1 |
| `ambiguous-short-date` | 1 | 1 | +0 |
| `invalid-input` | 3 | 1 | +2 |
| `non-iso-date-reject` | 3 | 1 | +2 |
| `bp-add-two` | 3 | 2 | +1 |
| `bp-latest-only` | 3 | 2 | +1 |
| `bp-history-limit-two` | 3 | 2 | +1 |
| `bp-bounded-range` | 3 | 2 | +1 |
| `bp-bounded-range-natural` | 3 | 4 | -1 |
| `bp-invalid-input` | 3 | 3 | +0 |
| `bp-non-iso-date-reject` | 4 | 2 | +2 |

## Metric Evidence

- `cli/add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb && go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb && go run ./cmd/openhealth weight list --limit 25'`.
- `cli/repeat-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb && go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb && go run ./cmd/openhealth weight list --limit 25'`.
- `cli/update-existing` broad repo search: `/bin/zsh -lc "pwd && sed -n '1,220p' <run-root>/cli/update-existing/repo/.agents/skills/openhealth/SKILL.md && printf '\\n---FILES---\\n' && rg --files . | sed -n '1,200p'"`.
- `cli/update-existing` generated path from broad search: `/bin/zsh -lc "pwd && sed -n '1,220p' <run-root>/cli/update-existing/repo/.agents/skills/openhealth/SKILL.md && printf '\\n---FILES---\\n' && rg --files . | sed -n '1,200p'"`.
- `cli/update-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 151.6 --unit lb'`.
- `cli/bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 2'`.
- `cli/bp-add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64 && go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 118 --diastolic 76 && go run ./cmd/openhealth blood-pressure list --limit 25'`.
- `cli/bp-latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/bp-history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 2'`.
- `cli/bp-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-bounded-range-natural` broad repo search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth repo . | sed -n '1,120p'"`; `/bin/zsh -lc "env | rg '"'^OPENHEALTH_'"'"`.
- `cli/bp-bounded-range-natural` generated path from broad search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth repo . | sed -n '1,120p'"`.
- `cli/bp-bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-invalid-input` broad repo search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && printf '\\n---\\n' && rg --files . | sed -n '1,200p'"`.
- `cli/bp-invalid-input` generated path from broad search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && printf '\\n---\\n' && rg --files . | sed -n '1,200p'"`.
- `cli/bp-invalid-input` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-31 --systolic 0 --diastolic -5 --pulse 0'`.
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
