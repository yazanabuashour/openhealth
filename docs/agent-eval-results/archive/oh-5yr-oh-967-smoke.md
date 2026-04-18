# oh-5yr Agent Eval Results

Date: oh-967-smoke

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `4`

Harness elapsed seconds: `83.32`

Reduced JSON artifact: `docs/agent-eval-results/archive/oh-5yr-oh-967-smoke.json`

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
| `production` | `add-two` | pass | pass | pass | 4 | 5 | 33.21 | in 186867 / cached 180224 / out 1160 | no | no | no | no | no | no |
| `production` | `latest-only` | pass | pass | pass | 2 | 4 | 28.00 | in 135515 / cached 129792 / out 524 | no | no | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 2 | 3 | 8.84 | in 67540 / cached 62592 / out 631 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 3 | 5 | 33.54 | in 139902 / cached 133376 / out 948 | no | no | no | no | no | no |
| `production` | `bp-latest-only` | pass | pass | pass | 3 | 3 | 19.73 | in 91297 / cached 86016 / out 475 | no | no | no | no | no | no |
| `production` | `bp-invalid-input` | pass | pass | pass | 2 | 3 | 11.14 | in 67645 / cached 60032 / out 602 | no | no | no | no | no | no |

## Comparison

Baseline: `docs/agent-eval-results/archive/oh-5yr-oh-967-smoke.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | +1 | +1 | +8.27 | -4250 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `latest-only` | same_pass | +0 | +0 | +5.88 | +964 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | +0 | +0 | -3.97 | +398 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | same_pass | -2 | -2 | -0.07 | -2569 | same_no | improved_to_no | same_no | same_no | same_no | same_no |
| `production` | `bp-latest-only` | same_pass | -2 | -4 | -4.81 | -6705 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-invalid-input` | same_pass | +1 | +1 | +3.62 | +1076 | same_no | same_no | same_no | same_no | same_no | same_no |

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
| `candidate_passes_all_scenarios` | pass | 6/6 candidate scenarios passed |
| `no_direct_generated_file_inspection` | pass | production must not directly inspect generated files |
| `no_module_cache_inspection` | pass | production must not inspect the Go module cache |
| `no_routine_broad_repo_search` | pass | production must not use broad repo search in routine scenarios |
| `no_openhealth_cli_usage` | pass | production must not use the openhealth CLI |
| `no_direct_sqlite_access` | pass | production must not use direct SQLite access |
| `total_tools_less_than_or_equal_cli` | fail | production tools 16 vs cli tools 0 |
| `minimum_scenarios_at_or_below_cli` | fail | 0 scenarios at or below CLI tools; required 5 of 6 |
| `no_routine_scenario_exceeds_cli_by_more_than_one_tool` | fail | missing cli scenarios: add-two, bp-add-two, bp-invalid-input, bp-latest-only, invalid-input, latest-only |

| Scenario | Candidate | CLI | Tools Δ |
| --- | ---: | ---: | ---: |
| `add-two` | 4 | 0 | n/a |
| `latest-only` | 2 | 0 | n/a |
| `invalid-input` | 2 | 0 | n/a |
| `bp-add-two` | 3 | 0 | n/a |
| `bp-latest-only` | 3 | 0 | n/a |
| `bp-invalid-input` | 2 | 0 | n/a |

## Scenario Notes

- `production/add-two`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/add-two/events.jsonl`.
- `production/latest-only`: expected unchanged seed rows and assistant output limited to latest row 2026-03-30; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/latest-only/events.jsonl`.
- `production/invalid-input`: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/production/invalid-input/events.jsonl`.
- `production/bp-add-two`: expected exactly two newest-first blood-pressure rows; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/bp-add-two/events.jsonl`.
- `production/bp-latest-only`: expected unchanged seed rows and assistant output limited to latest blood-pressure row 2026-03-30; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64, 2026-03-28 124/80] Raw event reference: `<run-root>/production/bp-latest-only/events.jsonl`.
- `production/bp-invalid-input`: expected no write and an invalid blood-pressure rejection; observed [] Raw event reference: `<run-root>/production/bp-invalid-input/events.jsonl`.

## CLI-Oriented Variant

Status: `runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands with a prewarmed per-scenario module cache`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
