# oh-5yr Agent Eval Results

Date: oh-967-final-r2

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `4`

Harness elapsed seconds: `342.76`

Reduced JSON artifact: `docs/agent-eval-results/archive/oh-5yr-oh-967-final-r2.json`

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
| `production` | `add-two` | pass | pass | pass | 2 | 4 | 35.14 | in 117043 / cached 105856 / out 653 | no | no | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 2 | 4 | 29.25 | in 138139 / cached 127744 / out 655 | no | no | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 2 | 5 | 33.91 | in 161875 / cached 155776 / out 835 | no | no | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 2 | 4 | 31.67 | in 117085 / cached 110464 / out 582 | no | no | no | no | no | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 2 | 3 | 23.99 | in 67894 / cached 63104 / out 507 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 2 | 4 | 21.13 | in 90871 / cached 86016 / out 425 | no | no | no | no | no | no |
| `production` | `history-limit-two` | pass | pass | pass | 2 | 4 | 20.05 | in 90950 / cached 86016 / out 458 | no | no | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 8.96 | in 44617 / cached 40704 / out 290 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 1 | 2 | 8.22 | in 45144 / cached 40192 / out 448 | no | yes | yes | no | no | no |
| `production` | `non-iso-date-reject` | pass | pass | pass | 1 | 2 | 7.15 | in 44620 / cached 40192 / out 304 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 2 | 4 | 23.86 | in 138612 / cached 127744 / out 757 | no | no | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 2 | 4 | 20.40 | in 114368 / cached 104832 / out 518 | no | no | no | no | no | no |
| `production` | `bp-history-limit-two` | pass | pass | pass | 2 | 4 | 24.94 | in 114333 / cached 108928 / out 504 | no | no | no | no | no | no |
| `production` | `bp-bounded-range` | pass | pass | pass | 2 | 4 | 22.85 | in 114813 / cached 108928 / out 583 | no | no | no | no | no | no |
| `production` | `bp-bounded-range-natural` | pass | pass | pass | 2 | 4 | 19.65 | in 114588 / cached 104832 / out 621 | no | no | no | no | no | no |
| `production` | `bp-invalid-input` | pass | pass | pass | 1 | 2 | 8.74 | in 45161 / cached 40192 / out 589 | no | no | no | no | no | no |
| `production` | `bp-non-iso-date-reject` | pass | pass | pass | 1 | 2 | 7.36 | in 45243 / cached 40192 / out 351 | no | no | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 4 | 4 | 24.32 | in 158978 / cached 153216 / out 788 | no | no | no | no | yes | no |
| `cli` | `repeat-add` | pass | pass | pass | 4 | 6 | 30.02 | in 136899 / cached 132352 / out 975 | no | no | no | no | yes | no |
| `cli` | `update-existing` | pass | pass | pass | 7 | 6 | 40.58 | in 259155 / cached 235904 / out 1379 | no | yes | yes | no | yes | no |
| `cli` | `bounded-range` | pass | pass | pass | 2 | 4 | 27.47 | in 89741 / cached 84992 / out 451 | no | no | no | no | yes | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 5 | 20.62 | in 112459 / cached 107392 / out 542 | no | no | no | no | yes | no |
| `cli` | `latest-only` | pass | pass | pass | 2 | 4 | 29.42 | in 89408 / cached 84992 / out 444 | no | no | no | no | yes | no |
| `cli` | `history-limit-two` | pass | pass | pass | 2 | 4 | 25.32 | in 114563 / cached 108928 / out 469 | no | no | no | no | yes | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 10.21 | in 44146 / cached 40192 / out 280 | no | no | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 1 | 2 | 11.36 | in 44075 / cached 40704 / out 554 | no | no | no | no | no | no |
| `cli` | `non-iso-date-reject` | pass | pass | pass | 1 | 2 | 7.35 | in 44130 / cached 37632 / out 456 | no | no | no | no | no | no |
| `cli` | `bp-add-two` | pass | pass | pass | 4 | 5 | 38.00 | in 183289 / cached 174592 / out 1060 | no | no | no | no | yes | no |
| `cli` | `bp-latest-only` | pass | pass | pass | 2 | 4 | 34.13 | in 112286 / cached 107904 / out 476 | no | no | no | no | yes | no |
| `cli` | `bp-history-limit-two` | pass | pass | pass | 2 | 4 | 23.96 | in 89465 / cached 84992 / out 441 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range` | pass | pass | pass | 4 | 7 | 36.05 | in 159884 / cached 149632 / out 1096 | no | no | no | no | yes | no |
| `cli` | `bp-bounded-range-natural` | pass | pass | pass | 2 | 3 | 17.68 | in 66838 / cached 62592 / out 399 | no | no | no | no | yes | no |
| `cli` | `bp-invalid-input` | pass | pass | pass | 8 | 6 | 30.70 | in 246640 / cached 222336 / out 1562 | yes | no | no | no | yes | no |
| `cli` | `bp-non-iso-date-reject` | pass | pass | pass | 2 | 2 | 8.90 | in 44890 / cached 40704 / out 699 | no | yes | yes | no | no | no |

## Comparison

Baseline: `docs/agent-eval-results/archive/oh-5yr-agentops-bp-final-r3.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | -1 | -1 | +10.68 | +4489 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `repeat-add` | same_pass | -1 | -1 | +4.71 | +3944 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `update-existing` | same_pass | -1 | +1 | +12.36 | -262 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range` | same_pass | -1 | -1 | +9.35 | +277 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bounded-range-natural` | same_pass | -1 | -2 | +2.25 | -1637 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | -1 | +1 | -2.14 | -962 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `history-limit-two` | same_pass | -1 | +0 | -0.38 | -544 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `ambiguous-short-date` | same_pass | +0 | +0 | +1.35 | -280 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | -2 | -3 | -16.28 | -1304 | same_no | regressed_to_yes | regressed_to_yes | same_no | same_no | same_no |
| `production` | `non-iso-date-reject` | same_pass | -2 | -3 | -17.15 | -6897 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | same_pass | -1 | +0 | -2.80 | -3269 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-latest-only` | same_pass | -1 | +0 | +0.27 | +4014 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-history-limit-two` | same_pass | -1 | -2 | +1.90 | -812 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range` | same_pass | -1 | +0 | +1.80 | +103 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-bounded-range-natural` | same_pass | -1 | +0 | -9.11 | +4051 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-invalid-input` | same_pass | -2 | -3 | -13.74 | -1627 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-non-iso-date-reject` | same_pass | -2 | -3 | -15.96 | -1392 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `add-two` | same_pass | +0 | -2 | -5.75 | -158 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `repeat-add` | same_pass | +0 | +2 | +6.56 | -1058 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `update-existing` | same_pass | +0 | +1 | +15.58 | +10599 | same_no | same_yes | regressed_to_yes | same_no | same_yes | same_no |
| `cli` | `bounded-range` | same_pass | +0 | -1 | +10.63 | -818 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bounded-range-natural` | same_pass | +0 | +2 | +3.52 | +484 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `latest-only` | same_pass | +0 | -1 | +13.07 | -510 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `history-limit-two` | same_pass | +0 | +0 | +4.08 | +525 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `ambiguous-short-date` | same_pass | +0 | +0 | +3.92 | -11 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | +0 | +0 | +1.74 | -501 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `non-iso-date-reject` | same_pass | +0 | +0 | -2.53 | +2574 | same_no | same_no | same_no | same_no | same_no | same_no |
| `cli` | `bp-add-two` | fixed | -1 | +0 | +7.23 | +2379 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-latest-only` | same_pass | +0 | +1 | +15.55 | -2323 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-history-limit-two` | same_pass | +0 | +1 | +7.70 | +299 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range` | same_pass | +2 | +3 | +17.33 | +4145 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-bounded-range-natural` | same_pass | +0 | -1 | -0.07 | -771 | same_no | same_no | same_no | same_no | same_yes | same_no |
| `cli` | `bp-invalid-input` | same_pass | +7 | +4 | +21.06 | +20428 | regressed_to_yes | same_no | same_no | same_no | regressed_to_yes | same_no |
| `cli` | `bp-non-iso-date-reject` | same_pass | +0 | -2 | -10.85 | -805 | same_no | regressed_to_yes | regressed_to_yes | same_no | improved_to_no | same_no |

## Metric Notes

- Production generated paths surfaced from broad repo searches in invalid-input in the oh-967-final-r2 run; this is tracked separately from direct generated-file inspection.
- Production broad repo search remained in invalid-input in the oh-967-final-r2 run.

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
| `candidate_passes_all_scenarios` | pass | 17/17 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `total_tools_less_than_or_equal_cli` | pass | production tools 29 vs cli tools 50 |
| `minimum_scenarios_at_or_below_cli` | pass | 17 scenarios at or below CLI tools; required 14 of 17 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 2 | 4 | -2 |
| `repeat-add` | 2 | 4 | -2 |
| `update-existing` | 2 | 7 | -5 |
| `bounded-range` | 2 | 2 | +0 |
| `bounded-range-natural` | 2 | 2 | +0 |
| `latest-only` | 2 | 2 | +0 |
| `history-limit-two` | 2 | 2 | +0 |
| `ambiguous-short-date` | 1 | 1 | +0 |
| `invalid-input` | 1 | 1 | +0 |
| `non-iso-date-reject` | 1 | 1 | +0 |
| `bp-add-two` | 2 | 4 | -2 |
| `bp-latest-only` | 2 | 2 | +0 |
| `bp-history-limit-two` | 2 | 2 | +0 |
| `bp-bounded-range` | 2 | 4 | -2 |
| `bp-bounded-range-natural` | 2 | 2 | +0 |
| `bp-invalid-input` | 1 | 8 | -7 |
| `bp-non-iso-date-reject` | 1 | 2 | -1 |

## Metric Evidence

- `production/invalid-input` broad repo search: `/bin/zsh -lc "pwd && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && printf '\\n---FILES---\\n' && rg --files . | sed -n '1,200p'"`.
- `production/invalid-input` generated path from broad search: `/bin/zsh -lc "pwd && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && printf '\\n---FILES---\\n' && rg --files . | sed -n '1,200p'"`.
- `cli/add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/repeat-add` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-29 --value 152.2 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight add --date 2026-03-30 --value 151.6 --unit lb'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 25'`.
- `cli/update-existing` broad repo search: `/bin/zsh -lc "rg -n \"weight add|duplicate|upsert|replace|INSERT|ON CONFLICT|weight\" cmd internal . -g '"'!**/vendor/**'"'"`.
- `cli/update-existing` generated path from broad search: `/bin/zsh -lc "rg -n \"weight add|duplicate|upsert|replace|INSERT|ON CONFLICT|weight\" cmd internal . -g '"'!**/vendor/**'"'"`.
- `cli/update-existing` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight --help'`; `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --from 2026-03-29 --to 2026-03-30'`.
- `cli/latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 1'`.
- `cli/history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth weight list --limit 2'`.
- `cli/bp-add-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-29 --systolic 122 --diastolic 78 --pulse 64'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-30 --systolic 118 --diastolic 76'`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-latest-only` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 1'`.
- `cli/bp-history-limit-two` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --limit 2'`.
- `cli/bp-bounded-range` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`; `/bin/zsh -lc "env | grep OPENHEALTH_DATABASE_PATH || true && go run ./cmd/openhealth --help | sed -n '1,220p'"`; `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list -db "$OPENHEALTH_DATABASE_PATH" --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-bounded-range-natural` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure list --from 2026-03-29 --to 2026-03-30'`.
- `cli/bp-invalid-input` direct generated-file: `/bin/zsh -lc "rg -n \"blood-pressure|parseDateOnly|systolic must be greater than 0|pulse must be greater than 0|diastolic must be greater than 0\" cmd internal client agentops -g'*.go'"`.
- `cli/bp-invalid-input` openhealth CLI: `/bin/zsh -lc 'go run ./cmd/openhealth blood-pressure add --date 2026-03-31 --systolic 0 --diastolic -5 --pulse 0'`.
- `cli/bp-non-iso-date-reject` broad repo search: `/bin/zsh -lc "pwd && rg --files .agents repo . | sed -n '1,120p'"`.
- `cli/bp-non-iso-date-reject` generated path from broad search: `/bin/zsh -lc "pwd && rg --files .agents repo . | sed -n '1,120p'"`.

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
