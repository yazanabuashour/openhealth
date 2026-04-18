# oh-5yr Agent Eval Results

Date: agentops-bp-baseline-r1

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/archive/oh-5yr-agentops-bp-baseline-r1.json`

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
| `production` | `add-two` | pass | pass | pass | 3 | 4 | 23.10 | in 140243 / cached 133888 / out 1238 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 4 | 5 | 23.07 | in 188676 / cached 181760 / out 1303 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 8 | 6 | 37.10 | in 303813 / cached 290432 / out 2635 | no | yes | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 3 | 3 | 20.72 | in 91675 / cached 86016 / out 844 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 3 | 4 | 20.23 | in 91574 / cached 86016 / out 888 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 2 | 3 | 20.39 | in 67519 / cached 44160 / out 667 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 4 | 3 | 22.61 | in 91590 / cached 86016 / out 980 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 6.37 | in 44450 / cached 40704 / out 328 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 3 | 4 | 27.43 | in 91802 / cached 86528 / out 1376 | no | no | no | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 3 | 3 | 22.27 | in 91691 / cached 86016 / out 1097 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 6 | 6 | 37.70 | in 238373 / cached 226176 / out 2153 | no | yes | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 3 | 4 | 21.17 | in 91611 / cached 86016 / out 846 | no | no | no | no | no | no |
| `production` | `bp-history-limit-two` | pass | pass | pass | 3 | 5 | 26.18 | in 140018 / cached 133888 / out 1036 | no | no | no | no | no | no |
| `production` | `bp-bounded-range` | pass | pass | pass | 3 | 5 | 23.54 | in 143197 / cached 136448 / out 1012 | no | no | no | no | no | no |
| `production` | `bp-bounded-range-natural` | fail | pass | fail | 3 | 4 | 26.08 | in 91799 / cached 86016 / out 929 | no | no | no | no | no | no |
| `production` | `bp-invalid-input` | pass | pass | pass | 3 | 5 | 22.80 | in 140348 / cached 119552 / out 1250 | no | no | no | no | no | no |
| `production` | `bp-non-iso-date-reject` | pass | pass | pass | 3 | 5 | 21.52 | in 140515 / cached 134400 / out 1198 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 6 | 7 | 33.93 | in 180209 / cached 170624 / out 1384 | no | yes | yes | no | yes | no |
| `cli` | `repeat-add` | pass | pass | pass | 10 | 7 | 34.21 | in 188502 / cached 173696 / out 1512 | no | yes | yes | no | yes | no |
| `cli` | `update-existing` | pass | pass | pass | 14 | 8 | 37.93 | in 306466 / cached 289536 / out 2342 | no | yes | yes | no | yes | no |
| `cli` | `bounded-range` | pass | pass | pass | 8 | 6 | 32.54 | in 210698 / cached 184320 / out 1445 | no | no | no | no | yes | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 9 | 6 | 34.43 | in 166076 / cached 152832 / out 1389 | no | yes | yes | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 4 | 4 | 17.44 | in 96123 / cached 88576 / out 774 | no | yes | no | no | yes | no |
| `cli` | `history-limit-two` | pass | pass | pass | 7 | 6 | 28.23 | in 189704 / cached 178304 / out 1209 | no | yes | yes | no | yes | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 6.87 | in 44093 / cached 40192 / out 307 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 1 | 2 | 5.92 | in 44022 / cached 40192 / out 311 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 2 | 4 | 20.11 | in 89562 / cached 84992 / out 510 | no | no | no | no | yes | no |
| `cli` | `bp-add-two` | pass | pass | pass | 4 | 4 | 25.51 | in 136173 / cached 130816 / out 859 | no | no | no | no | yes | no |
| `cli` | `bp-latest-only` | pass | pass | pass | 7 | 5 | 27.37 | in 161458 / cached 145664 / out 1186 | no | yes | yes | no | yes | no |
| `cli` | `bp-history-limit-two` | pass | pass | pass | 8 | 5 | 26.31 | in 182426 / cached 172160 / out 1348 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range` | pass | pass | pass | 6 | 5 | 24.35 | in 181863 / cached 171136 / out 1043 | no | yes | no | no | yes | no |
| `cli` | `bp-bounded-range-natural` | pass | pass | pass | 16 | 9 | 86.12 | in 538492 / cached 519040 / out 3360 | no | yes | yes | no | yes | no |
| `cli` | `bp-invalid-input` | pass | pass | pass | 2 | 4 | 22.69 | in 112910 / cached 107904 / out 826 | no | no | no | no | yes | no |
| `cli` | `bp-non-iso-date-reject` | pass | pass | pass | 3 | 4 | 23.05 | in 136370 / cached 126720 / out 936 | no | no | no | no | yes | no |

## Comparison

Baseline: `docs/agent-eval-results/archive/oh-5yr-agentops-production-expanded-r3.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | +1 | +1 | +2.33 | +506 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `repeat-add` | same_pass | +2 | +2 | +2.54 | +3294 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `update-existing` | same_pass | +6 | +2 | +18.04 | +9455 | same_no | regressed_to_yes | same_no | same_no | same_no | same_no |
| `production` | `bounded-range` | same_pass | +1 | -1 | +0.90 | -817 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range-natural` | same_pass | +1 | +0 | -1.60 | -1094 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | +0 | +0 | +1.33 | +20038 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `history-limit-two` | same_pass | +2 | -1 | +2.50 | -1106 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `ambiguous-short-date` | same_pass | +1 | +1 | +2.05 | -254 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | +2 | +2 | +19.18 | +2375 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | same_pass | +2 | +1 | +12.36 | +105 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-latest-only` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-history-limit-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-bounded-range-natural` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-invalid-input` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `production` | `bp-non-iso-date-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `add-two` | same_pass | -2 | +1 | -15.99 | +567 | same_no | regressed_to_yes | regressed_to_yes | same_no | regressed_to_yes | same_no |
| `cli` | `repeat-add` | same_pass | +0 | +0 | -20.90 | +1810 | same_no | regressed_to_yes | regressed_to_yes | same_no | regressed_to_yes | same_no |
| `cli` | `update-existing` | same_pass | +4 | +3 | -17.15 | +6065 | same_no | regressed_to_yes | regressed_to_yes | same_no | regressed_to_yes | same_no |
| `cli` | `bounded-range` | same_pass | -3 | -1 | -41.48 | -247 | same_no | same_no | same_no | same_no | regressed_to_yes | same_no |
| `cli` | `bounded-range-natural` | same_pass | -3 | -4 | -72.14 | -1277 | same_no | regressed_to_yes | regressed_to_yes | same_no | same_yes | same_no |
| `cli` | `latest-only` | same_pass | +1 | -1 | -22.57 | +1502 | same_no | regressed_to_yes | same_no | same_no | same_yes | same_no |
| `cli` | `history-limit-two` | same_pass | -11 | +1 | -36.62 | -2415 | same_no | regressed_to_yes | regressed_to_yes | same_no | regressed_to_yes | same_no |
| `cli` | `ambiguous-short-date` | same_pass | +1 | +1 | +1.83 | -80 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | +0 | +0 | -1.52 | -980 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `non-iso-date-reject` | same_pass | +1 | +2 | +11.51 | -332 | same_no | same_no | same_no | same_no | regressed_to_yes | same_no |
| `cli` | `bp-add-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-latest-only` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-history-limit-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-bounded-range-natural` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-invalid-input` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |
| `cli` | `bp-non-iso-date-reject` | new | n/a | n/a | n/a | n/a |  |  |  |  |  |  |

## Metric Notes

- Production broad repo search remained in update-existing, bp-add-two in the agentops-bp-baseline-r1 run.

## Production Stop-Loss

- Triggered: `yes`
- Recommendation: `continue_cli_baseline_for_agent_operations`
- Trigger: production correctness below 100% in bp-bounded-range-natural
- Trigger: broad repo search in more than one routine scenario: bp-add-two, update-existing

## Code-First CLI Comparison

- Candidate: `production`
- Baseline: `cli`
- Beats CLI: `no`
- Recommendation: `continue_cli_for_routine_openhealth_operations`

| Criterion | Result | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | fail | 16/17 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | fail | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `total_tools_less_than_or_equal_cli` | pass | production tools 58 vs cli tools 108 |
| `minimum_scenarios_at_or_below_cli` | fail | 13 scenarios at or below CLI tools; required 14 of 17 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | fail | routine scenarios over CLI by more than one tool: bp-add-two |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 3 | 6 | -3 |
| `repeat-add` | 4 | 10 | -6 |
| `update-existing` | 8 | 14 | -6 |
| `bounded-range` | 3 | 8 | -5 |
| `bounded-range-natural` | 3 | 9 | -6 |
| `latest-only` | 2 | 4 | -2 |
| `history-limit-two` | 4 | 7 | -3 |
| `ambiguous-short-date` | 1 | 1 | +0 |
| `invalid-input` | 3 | 1 | +2 |
| `non-iso-date-reject` | 3 | 2 | +1 |
| `bp-add-two` | 6 | 4 | +2 |
| `bp-latest-only` | 3 | 7 | -4 |
| `bp-history-limit-two` | 3 | 8 | -5 |
| `bp-bounded-range` | 3 | 6 | -3 |
| `bp-bounded-range-natural` | 3 | 16 | -13 |
| `bp-invalid-input` | 3 | 2 | +1 |
| `bp-non-iso-date-reject` | 3 | 3 | +0 |

## Metric Evidence

- `production/update-existing` broad repo search: `/bin/zsh -lc 'rg -n "RunWeightTask|WeightTaskRequest|WeightTaskActionUpsert|WeightInput|WeightListModeRange" .'`.
- `production/bp-add-two` broad repo search: `/bin/zsh -lc 'rg -n "type BloodPressureTaskRequest|RunBloodPressureTask|BloodPressureInput|BloodPressureListMode" .'`.
- `cli/add-two` broad repo search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && printf '\\n---\\n' && rg --files . | sed -n '1,120p'"`.
- `cli/add-two` generated path from broad search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && printf '\\n---\\n' && rg --files . | sed -n '1,120p'"`.
- `cli/add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 10'`.
- `cli/repeat-add` broad repo search: `/bin/zsh -lc 'rg --files .'`; `/bin/zsh -lc 'rg -n "data path|data_path|OPENHEALTH|sqlite|weight add|weight list|config" -S .'`.
- `cli/repeat-add` generated path from broad search: `/bin/zsh -lc 'rg --files .'`; `/bin/zsh -lc 'rg -n "data path|data_path|OPENHEALTH|sqlite|weight add|weight list|config" -S .'`.
- `cli/repeat-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth help'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/update-existing` broad repo search: `/bin/zsh -lc "pwd && sed -n '1,220p' <run-root>/cli/update-existing/repo/.agents/skills/openhealth/SKILL.md && printf '\\n---FILES---\\n' && rg --files . | sed -n '1,200p'"`.
- `cli/update-existing` generated path from broad search: `/bin/zsh -lc "pwd && sed -n '1,220p' <run-root>/cli/update-existing/repo/.agents/skills/openhealth/SKILL.md && printf '\\n---FILES---\\n' && rg --files . | sed -n '1,200p'"`.
- `cli/update-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --db <run-root>/cli/update-existing/openhealth.db --from 2026-03-29 --to 2026-03-29'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --db <run-root>/cli/update-existing/openhealth.db --date 2026-03-29 --value 151.6 --unit lb'`.
- `cli/bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list -db <run-root>/cli/bounded-range/openhealth.db --from 2026-03-29 --to 2026-03-30'`.
- `cli/bounded-range-natural` broad repo search: `/bin/zsh -lc "rg --files .agents . | sed -n '1,200p'"`; `/bin/zsh -lc "rg -n \"data path|datapath|local data|OPENHEALTH|sqlite|db\" .agents . -g '"'!**/node_modules/**'"'"`.
- `cli/bounded-range-natural` generated path from broad search: `/bin/zsh -lc "rg --files .agents . | sed -n '1,200p'"`.
- `cli/bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list -db <run-root>/cli/bounded-range-natural/openhealth.db --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` broad repo search: `/bin/zsh -lc 'rg -n "data path|data_path|OPENHEALTH|sqlite|db" -S .agents . | head -n 80'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 5'`.
- `cli/history-limit-two` broad repo search: `/bin/zsh -lc "rg -n \"data path|data_path|OPENHEALTH|sqlite|db\" .agents . -g '"'!**/node_modules/**'"'"`; `/bin/zsh -lc "rg --files .agents . | sed -n '1,200p'"`.
- `cli/history-limit-two` generated path from broad search: `/bin/zsh -lc "rg --files .agents . | sed -n '1,200p'"`.
- `cli/history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 2'`.
- `cli/non-iso-date-reject` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026/03/31 --value 152.2 --unit lb'`.
- `cli/bp-add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 118 --diastolic 76'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-latest-only` broad repo search: `/bin/zsh -lc 'rg --files .'`; `/bin/zsh -lc 'rg -n "data path|data_path|local data|sqlite|db path|openhealth" -S .'`.
- `cli/bp-latest-only` generated path from broad search: `/bin/zsh -lc 'rg --files .'`; `/bin/zsh -lc 'rg -n "data path|data_path|local data|sqlite|db path|openhealth" -S .'`.
- `cli/bp-latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/bp-history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 2'`.
- `cli/bp-bounded-range` broad repo search: `/bin/zsh -lc "rg -n \"data path|datapath|local data|OPENHEALTH|blood-pressure list|from --to|sqlite\" .agents . -g '"'!**/node_modules/**'"'"`.
- `cli/bp-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-bounded-range-natural` broad repo search: `/bin/zsh -lc 'rg -n "data path|DATA_PATH|openhealth|sqlite|db" -S .'`; `/bin/zsh -lc 'rg --files .'`.
- `cli/bp-bounded-range-natural` generated path from broad search: `/bin/zsh -lc 'rg -n "data path|DATA_PATH|openhealth|sqlite|db" -S .'`; `/bin/zsh -lc 'rg --files .'`.
- `cli/bp-bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list -db <run-root>/cli/bp-bounded-range-natural/openhealth.db --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-invalid-input` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-31 --systolic 0 --diastolic -5 --pulse 0'`.
- `cli/bp-non-iso-date-reject` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026/03/31 --systolic 122 --diastolic 78'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026/03/31 --systolic 122 --diastolic 78; echo EXIT:$?'`.

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
