# oh-5yr Agent Eval Results

Date: code-first-iter1

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/archive/oh-5yr-code-first-iter1.json`

Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.

## History Isolation

- Status: `review`
- Every agent run used `codex exec --ephemeral` from `<run-root>/<variant>/<scenario>/repo`.
- The Codex desktop app was not used for eval prompts.
- New Codex session files referencing `<run-root>`: `1`.
- Limitation: A session-file count changed while evals ran; this may be from another Codex process, because the harness uses --ephemeral and a throwaway cwd.

## Results

| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- |
| `agentops-code` | `add-two` | pass | pass | pass | 7 | 8 | 52.62 | in 210323 / cached 196608 / out 4108 | no | no | no | no |
| `agentops-code` | `repeat-add` | pass | pass | pass | 15 | 9 | 48.45 | in 297028 / cached 279424 / out 4816 | no | yes | yes | no |
| `agentops-code` | `update-existing` | pass | pass | pass | 26 | 12 | 77.11 | in 629267 / cached 604288 / out 7339 | yes | yes | no | yes |
| `agentops-code` | `bounded-range` | pass | pass | pass | 10 | 8 | 47.11 | in 308777 / cached 267392 / out 3935 | yes | no | no | no |
| `agentops-code` | `bounded-range-natural` | pass | pass | pass | 11 | 7 | 46.41 | in 203584 / cached 173696 / out 3298 | no | yes | no | no |
| `agentops-code` | `ambiguous-short-date` | pass | pass | pass | 2 | 3 | 10.07 | in 93278 / cached 75904 / out 805 | no | yes | yes | no |
| `agentops-code` | `invalid-input` | pass | pass | pass | 1 | 2 | 9.00 | in 53625 / cached 40704 / out 415 | no | yes | yes | no |
| `cli` | `add-two` | pass | pass | pass | 4 | 4 | 28.92 | in 139413 / cached 133376 / out 818 | no | no | no | no |
| `cli` | `repeat-add` | pass | pass | pass | 2 | 5 | 23.58 | in 118211 / cached 106880 / out 745 | no | no | no | no |
| `cli` | `update-existing` | pass | pass | pass | 5 | 5 | 31.94 | in 187015 / cached 175616 / out 1190 | no | no | no | no |
| `cli` | `bounded-range` | pass | pass | pass | 4 | 6 | 34.07 | in 227295 / cached 217984 / out 1787 | no | yes | no | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 4 | 21.78 | in 115188 / cached 106368 / out 523 | no | no | no | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 2 | 2 | 7.14 | in 45843 / cached 40704 / out 440 | no | yes | yes | no |
| `cli` | `invalid-input` | pass | pass | pass | 3 | 3 | 12.08 | in 70037 / cached 63104 / out 792 | no | yes | yes | no |

## Comparison

Baseline: `docs/agent-eval-results/archive/oh-5yr-2026-04-17-production-hardening.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- |
| `agentops-code` | `add-two` | new | n/a | n/a | n/a | n/a |  |  |  |  |
| `agentops-code` | `repeat-add` | new | n/a | n/a | n/a | n/a |  |  |  |  |
| `agentops-code` | `update-existing` | new | n/a | n/a | n/a | n/a |  |  |  |  |
| `agentops-code` | `bounded-range` | new | n/a | n/a | n/a | n/a |  |  |  |  |
| `agentops-code` | `bounded-range-natural` | new | n/a | n/a | n/a | n/a |  |  |  |  |
| `agentops-code` | `ambiguous-short-date` | new | n/a | n/a | n/a | n/a |  |  |  |  |
| `agentops-code` | `invalid-input` | new | n/a | n/a | n/a | n/a |  |  |  |  |
| `cli` | `add-two` | same_pass | -2 | -1 | -4.52 | -444 | same_no | same_no | same_no | same_no |
| `cli` | `repeat-add` | same_pass | -2 | -1 | -311.86 | -14784 | same_no | same_no | same_no | same_no |
| `cli` | `update-existing` | same_pass | +3 | +1 | +8.86 | +5603 | same_no | same_no | same_no | same_no |
| `cli` | `bounded-range` | same_pass | +1 | +2 | +12.03 | +3507 | same_no | regressed_to_yes | same_no | same_no |
| `cli` | `bounded-range-natural` | same_pass | +0 | -1 | -1.76 | +3007 | same_no | same_no | same_no | same_no |
| `cli` | `ambiguous-short-date` | same_pass | -2 | -1 | -2.60 | -2862 | same_no | same_yes | regressed_to_yes | same_no |
| `cli` | `invalid-input` | same_pass | +2 | +1 | +3.12 | +2501 | same_no | regressed_to_yes | regressed_to_yes | same_no |

## Code-First CLI Comparison

- Candidate: `agentops-code`
- Baseline: `cli`
- Beats CLI: `no`
- Recommendation: `continue_cli_for_routine_weight_operations`

| Criterion | Result | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | pass | 7/7 candidate scenarios present |
| `no_direct_generated_file_inspection` | fail | agentops-code must not directly inspect generated files |
| `no_module_cache_inspection` | fail | agentops-code must not inspect the Go module cache |
| `no_routine_broad_repo_search` | fail | agentops-code must not use broad repo search in routine weight scenarios |
| `total_tools_less_than_or_equal_cli` | fail | agentops-code tools 72 vs cli tools 22 |
| `at_least_five_scenarios_at_or_below_cli` | fail | 2 scenarios at or below CLI tools |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | fail | routine scenarios over CLI by more than one tool: add-two, bounded-range, bounded-range-natural, repeat-add, update-existing |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 7 | 4 | +3 |
| `repeat-add` | 15 | 2 | +13 |
| `update-existing` | 26 | 5 | +21 |
| `bounded-range` | 10 | 4 | +6 |
| `bounded-range-natural` | 11 | 2 | +9 |
| `ambiguous-short-date` | 2 | 2 | +0 |
| `invalid-input` | 1 | 3 | -2 |

## Metric Evidence

- `agentops-code/repeat-add` broad repo search: `/bin/zsh -lc 'rg -n "weight|weights|OpenHealth|local data path|configured local data path" -S .'`.
- `agentops-code/repeat-add` generated path from broad search: `/bin/zsh -lc 'rg -n "weight|weights|OpenHealth|local data path|configured local data path" -S .'`.
- `agentops-code/update-existing` direct generated-file: `/bin/zsh -lc "rg --files client internal agentops | rg 'local|weight|config|store|data|json|yaml|toml'"`.
- `agentops-code/update-existing` broad repo search: `/bin/zsh -lc "git status --short && printf '\\n---\\n' && rg -n \"2026-03-29|152\\.2|151\\.6|weight\" ."`.
- `agentops-code/update-existing` module cache: `/bin/zsh -lc 'go env GOPATH GOMODCACHE GOPROXY GOSUMDB'`.
- `agentops-code/bounded-range` direct generated-file: `/bin/zsh -lc "rg -n \"LocalConfig|data path|sqlite|history|weight\" client agentops internal -g '"'!**/*_test.go'"'"`.
- `agentops-code/bounded-range-natural` broad repo search: `/bin/zsh -lc "pwd && bd prime && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && rg -n \"local data path|data path|weights|weight\" -S .agents repo . 2>/dev/null | head -n 80"`.
- `agentops-code/ambiguous-short-date` broad repo search: `/bin/zsh -lc "bd prime && printf '\\n---\\n' && sed -n '1,220p' <run-root> && printf '\\n---\\n' && rg -n \"weight|OpenHealth|03/29|ambiguous short date|date\" -S ."`.
- `agentops-code/ambiguous-short-date` generated path from broad search: `/bin/zsh -lc "bd prime && printf '\\n---\\n' && sed -n '1,220p' <run-root> && printf '\\n---\\n' && rg -n \"weight|OpenHealth|03/29|ambiguous short date|date\" -S ."`.
- `agentops-code/invalid-input` broad repo search: `/bin/zsh -lc "pwd && bd prime && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && rg -n \"weight|stone|OpenHealth|entries|log\" -S . --hidden --glob '"'!.git'"'"`.
- `agentops-code/invalid-input` generated path from broad search: `/bin/zsh -lc "pwd && bd prime && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && rg -n \"weight|stone|OpenHealth|entries|log\" -S . --hidden --glob '"'!.git'"'"`.
- `cli/bounded-range` broad repo search: `/bin/zsh -lc "bd prime && pwd && rg --files -g 'SKILL.md' -g 'AGENTS.md' -g '.agents/**' -g 'bd*' -g '*openhealth*' ."`.
- `cli/ambiguous-short-date` broad repo search: `/bin/zsh -lc "rg --files . | sed -n '1,200p'"`.
- `cli/ambiguous-short-date` generated path from broad search: `/bin/zsh -lc "rg --files . | sed -n '1,200p'"`.
- `cli/invalid-input` broad repo search: `/bin/zsh -lc "pwd && bd prime && rg --files -g 'SKILL.md' -g 'AGENTS.md' -g '*.md' -g '*.csv' -g '*.tsv' -g '*.json' -g '*.yaml' -g '*.yml' ."`.
- `cli/invalid-input` generated path from broad search: `/bin/zsh -lc "pwd && bd prime && rg --files -g 'SKILL.md' -g 'AGENTS.md' -g '*.md' -g '*.csv' -g '*.tsv' -g '*.json' -g '*.yaml' -g '*.yml' ."`.

## Scenario Notes

- `agentops-code/add-two`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/agentops-code/add-two/events.jsonl`.
- `agentops-code/repeat-add`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/agentops-code/repeat-add/events.jsonl`.
- `agentops-code/update-existing`: expected one updated row; observed [2026-03-29 151.6 lb] Raw event reference: `<run-root>/agentops-code/update-existing/events.jsonl`.
- `agentops-code/bounded-range`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/agentops-code/bounded-range/events.jsonl`.
- `agentops-code/bounded-range-natural`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/agentops-code/bounded-range-natural/events.jsonl`.
- `agentops-code/ambiguous-short-date`: expected no write and a year clarification; observed [] Raw event reference: `<run-root>/agentops-code/ambiguous-short-date/events.jsonl`.
- `agentops-code/invalid-input`: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/agentops-code/invalid-input/events.jsonl`.
- `cli/add-two`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/cli/add-two/events.jsonl`.
- `cli/repeat-add`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/cli/repeat-add/events.jsonl`.
- `cli/update-existing`: expected one updated row; observed [2026-03-29 151.6 lb] Raw event reference: `<run-root>/cli/update-existing/events.jsonl`.
- `cli/bounded-range`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/cli/bounded-range/events.jsonl`.
- `cli/bounded-range-natural`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/cli/bounded-range-natural/events.jsonl`.
- `cli/ambiguous-short-date`: expected no write and a year clarification; observed [] Raw event reference: `<run-root>/cli/ambiguous-short-date/events.jsonl`.
- `cli/invalid-input`: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/cli/invalid-input/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight add/list with a prewarmed per-scenario module cache`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
