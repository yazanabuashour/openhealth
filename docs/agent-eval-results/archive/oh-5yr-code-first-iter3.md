# oh-5yr Agent Eval Results

Date: code-first-iter3

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-code-first-iter3.json`

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
| `agentops-code` | `add-two` | pass | pass | pass | 1 | 2 | 25.39 | in 46063 / cached 41728 / out 883 | no | no | no | no |
| `agentops-code` | `repeat-add` | pass | pass | pass | 2 | 4 | 28.65 | in 118511 / cached 105344 / out 1841 | no | no | no | no |
| `agentops-code` | `update-existing` | pass | pass | pass | 1 | 3 | 22.85 | in 69938 / cached 65664 / out 1127 | no | no | no | no |
| `agentops-code` | `bounded-range` | pass | pass | pass | 1 | 2 | 20.79 | in 46147 / cached 41728 / out 886 | no | no | no | no |
| `agentops-code` | `bounded-range-natural` | pass | pass | pass | 2 | 4 | 22.51 | in 94805 / cached 89088 / out 939 | no | no | no | no |
| `agentops-code` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 4.14 | in 22560 / cached 18816 / out 136 | no | no | no | no |
| `agentops-code` | `invalid-input` | pass | pass | pass | 0 | 1 | 5.14 | in 22538 / cached 18816 / out 168 | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 7 | 5 | 31.40 | in 196629 / cached 172032 / out 1171 | no | yes | no | no |
| `cli` | `repeat-add` | pass | pass | pass | 6 | 7 | 43.19 | in 280460 / cached 268032 / out 2734 | no | yes | no | no |
| `cli` | `update-existing` | pass | pass | pass | 8 | 6 | 43.19 | in 263379 / cached 245248 / out 1898 | no | yes | yes | no |
| `cli` | `bounded-range` | pass | pass | pass | 2 | 4 | 24.29 | in 115538 / cached 109952 / out 578 | no | no | no | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 4 | 19.72 | in 115209 / cached 109440 / out 537 | no | no | no | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 7.66 | in 45186 / cached 40704 / out 354 | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 2 | 2 | 8.98 | in 45239 / cached 40704 / out 752 | no | no | no | no |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-code-first-iter2.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- |
| `agentops-code` | `add-two` | same_pass | -4 | -5 | -45.56 | -2999 | same_no | same_no | same_no | same_no |
| `agentops-code` | `repeat-add` | same_pass | -2 | -2 | -13.36 | +7163 | same_no | same_no | same_no | same_no |
| `agentops-code` | `update-existing` | same_pass | -12 | -9 | -58.96 | -5455 | same_no | same_no | same_no | same_no |
| `agentops-code` | `bounded-range` | same_pass | -5 | -5 | -32.82 | -2725 | same_no | same_no | same_no | same_no |
| `agentops-code` | `bounded-range-natural` | fixed | -2 | -2 | -15.13 | -372 | same_no | same_no | same_no | same_no |
| `agentops-code` | `ambiguous-short-date` | same_pass | +0 | +0 | -0.93 | +97 | same_no | same_no | same_no | same_no |
| `agentops-code` | `invalid-input` | same_pass | +0 | +0 | -2.05 | +97 | same_no | same_no | same_no | same_no |
| `cli` | `add-two` | same_pass | +3 | +2 | -0.19 | +18125 | same_no | regressed_to_yes | same_no | same_no |
| `cli` | `repeat-add` | same_pass | -4 | +1 | +2.50 | +1529 | same_no | regressed_to_yes | same_no | same_no |
| `cli` | `update-existing` | same_pass | +6 | +2 | +21.37 | +12615 | same_no | regressed_to_yes | regressed_to_yes | same_no |
| `cli` | `bounded-range` | same_pass | +0 | +0 | +3.29 | +239 | same_no | same_no | same_no | same_no |
| `cli` | `bounded-range-natural` | same_pass | +0 | +0 | -3.26 | +559 | same_no | same_no | same_no | same_no |
| `cli` | `ambiguous-short-date` | same_pass | +0 | +0 | -2.16 | +21 | same_no | same_no | same_no | same_no |
| `cli` | `invalid-input` | same_pass | -3 | -1 | -3.23 | -4963 | same_no | improved_to_no | same_no | same_no |

## Code-First CLI Comparison

- Candidate: `agentops-code`
- Baseline: `cli`
- Beats CLI: `yes`
- Recommendation: `prefer_agentops_code_for_routine_weight_operations`

| Criterion | Result | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | pass | 7/7 candidate scenarios present |
| `no_direct_generated_file_inspection` | pass | agentops-code must not directly inspect generated files |
| `no_module_cache_inspection` | pass | agentops-code must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | agentops-code must not use broad repo search in routine weight scenarios |
| `total_tools_less_than_or_equal_cli` | pass | agentops-code tools 7 vs cli tools 28 |
| `at_least_five_scenarios_at_or_below_cli` | pass | 7 scenarios at or below CLI tools |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | pass | no routine scenario exceeded CLI by more than one tool |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 1 | 7 | -6 |
| `repeat-add` | 2 | 6 | -4 |
| `update-existing` | 1 | 8 | -7 |
| `bounded-range` | 1 | 2 | -1 |
| `bounded-range-natural` | 2 | 2 | +0 |
| `ambiguous-short-date` | 0 | 1 | -1 |
| `invalid-input` | 0 | 2 | -2 |

## Metric Evidence

- `cli/add-two` broad repo search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth repo . | rg 'SKILL\\.md"'$|openhealth|bd'"'"`.
- `cli/repeat-add` broad repo search: `/bin/zsh -lc "bd prime && pwd && rg --files -g 'SKILL.md' -g 'AGENTS.md' -g '*weight*' -g '*openhealth*' -g '*.json' -g '*.csv' -g '*.tsv' -g '*.md' ."`.
- `cli/update-existing` broad repo search: `/bin/zsh -lc "bd prime && printf '\\n---\\n' && rg -n \"03/29/2026|152\\.2|151\\.6|weight\" ."`.
- `cli/update-existing` generated path from broad search: `/bin/zsh -lc "bd prime && printf '\\n---\\n' && rg -n \"03/29/2026|152\\.2|151\\.6|weight\" ."`.

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
