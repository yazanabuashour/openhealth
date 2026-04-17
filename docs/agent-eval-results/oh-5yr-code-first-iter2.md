# oh-5yr Agent Eval Results

Date: code-first-iter2

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-code-first-iter2.json`

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
| `agentops-code` | `add-two` | pass | pass | pass | 5 | 7 | 70.95 | in 238118 / cached 230784 / out 5218 | no | no | no | no |
| `agentops-code` | `repeat-add` | pass | pass | pass | 4 | 6 | 42.01 | in 147572 / cached 141568 / out 3142 | no | no | no | no |
| `agentops-code` | `update-existing` | pass | pass | pass | 13 | 12 | 81.81 | in 342017 / cached 332288 / out 8908 | no | no | no | no |
| `agentops-code` | `bounded-range` | pass | pass | pass | 6 | 7 | 53.61 | in 175720 / cached 168576 / out 4635 | no | no | no | no |
| `agentops-code` | `bounded-range-natural` | fail | pass | fail | 4 | 6 | 37.64 | in 169033 / cached 162944 / out 2564 | no | no | no | no |
| `agentops-code` | `ambiguous-short-date` | pass | pass | pass | 0 | 1 | 5.07 | in 22463 / cached 18816 / out 224 | no | no | no | no |
| `agentops-code` | `invalid-input` | pass | pass | pass | 0 | 1 | 7.19 | in 22441 / cached 18816 / out 282 | no | no | no | no |
| `cli` | `add-two` | pass | pass | pass | 4 | 3 | 31.59 | in 164808 / cached 158336 / out 1199 | no | no | no | no |
| `cli` | `repeat-add` | pass | pass | pass | 10 | 6 | 40.69 | in 230931 / cached 220032 / out 1742 | no | no | no | no |
| `cli` | `update-existing` | pass | pass | pass | 2 | 4 | 21.82 | in 115468 / cached 109952 / out 629 | no | no | no | no |
| `cli` | `bounded-range` | pass | pass | pass | 2 | 4 | 21.00 | in 91875 / cached 86528 / out 461 | no | no | no | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 4 | 22.98 | in 91738 / cached 86528 / out 503 | no | no | no | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 9.82 | in 45165 / cached 40704 / out 348 | no | no | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 5 | 3 | 12.21 | in 72602 / cached 63104 / out 908 | no | yes | no | no |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-code-first-iter1.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- |
| `agentops-code` | `add-two` | same_pass | -2 | -1 | +18.33 | -6381 | same_no | same_no | same_no | same_no |
| `agentops-code` | `repeat-add` | same_pass | -11 | -3 | -6.44 | -11600 | same_no | improved_to_no | improved_to_no | same_no |
| `agentops-code` | `update-existing` | same_pass | -13 | +0 | +4.70 | -15250 | improved_to_no | improved_to_no | same_no | improved_to_no |
| `agentops-code` | `bounded-range` | same_pass | -4 | -1 | +6.50 | -34241 | improved_to_no | same_no | same_no | same_no |
| `agentops-code` | `bounded-range-natural` | regressed | -7 | -1 | -8.77 | -23799 | same_no | improved_to_no | same_no | same_no |
| `agentops-code` | `ambiguous-short-date` | same_pass | -2 | -2 | -5.00 | -13727 | same_no | improved_to_no | improved_to_no | same_no |
| `agentops-code` | `invalid-input` | same_pass | -1 | -1 | -1.81 | -9296 | same_no | improved_to_no | improved_to_no | same_no |
| `cli` | `add-two` | same_pass | +0 | -1 | +2.67 | +435 | same_no | same_no | same_no | same_no |
| `cli` | `repeat-add` | same_pass | +8 | +1 | +17.11 | -432 | same_no | same_no | same_no | same_no |
| `cli` | `update-existing` | same_pass | -3 | -1 | -10.12 | -5883 | same_no | same_no | same_no | same_no |
| `cli` | `bounded-range` | same_pass | -2 | -2 | -13.07 | -3964 | same_no | improved_to_no | same_no | same_no |
| `cli` | `bounded-range-natural` | same_pass | +0 | +0 | +1.20 | -3610 | same_no | same_no | same_no | same_no |
| `cli` | `ambiguous-short-date` | same_pass | -1 | +0 | +2.68 | -678 | same_no | improved_to_no | improved_to_no | same_no |
| `cli` | `invalid-input` | same_pass | +2 | +0 | +0.13 | +2565 | same_no | same_yes | improved_to_no | same_no |

## Code-First CLI Comparison

- Candidate: `agentops-code`
- Baseline: `cli`
- Beats CLI: `no`
- Recommendation: `continue_cli_for_routine_weight_operations`

| Criterion | Result | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | fail | 7/7 candidate scenarios present |
| `no_direct_generated_file_inspection` | pass | agentops-code must not directly inspect generated files |
| `no_module_cache_inspection` | pass | agentops-code must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | agentops-code must not use broad repo search in routine weight scenarios |
| `total_tools_less_than_or_equal_cli` | fail | agentops-code tools 32 vs cli tools 26 |
| `at_least_five_scenarios_at_or_below_cli` | fail | 3 scenarios at or below CLI tools |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | fail | routine scenarios over CLI by more than one tool: bounded-range, bounded-range-natural, update-existing |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 5 | 4 | +1 |
| `repeat-add` | 4 | 10 | -6 |
| `update-existing` | 13 | 2 | +11 |
| `bounded-range` | 6 | 2 | +4 |
| `bounded-range-natural` | 4 | 2 | +2 |
| `ambiguous-short-date` | 0 | 1 | -1 |
| `invalid-input` | 0 | 5 | -5 |

## Metric Evidence

- `cli/invalid-input` broad repo search: `/bin/zsh -lc "bd prime && printf '\\n---\\n' && pwd && printf '\\n---\\n' && rg --files -g 'SKILL.md' -g 'AGENTS.md' -g 'README*' -g '.bd*' -g 'openhealth*' -g '*weight*' -g '*weights*' ."`.

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
