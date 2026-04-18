# oh-5yr Agent Eval Results

Date: 2026-04-17-production-hardening

Harness: `codex exec --json --ephemeral --full-auto from throwaway run directories`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-2026-04-17-production-hardening.json`

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
| `production` | `add-two` | pass | pass | pass | 3 | 6 | 44.47 | in 204024 / cached 195072 / out 1885 | no | no | no | no |
| `production` | `repeat-add` | pass | pass | pass | 3 | 6 | 47.26 | in 231540 / cached 222592 / out 1902 | no | no | no | no |
| `production` | `update-existing` | pass | pass | pass | 14 | 10 | 63.20 | in 311006 / cached 294144 / out 3029 | no | no | no | no |
| `production` | `bounded-range` | pass | pass | pass | 7 | 5 | 36.72 | in 159048 / cached 148736 / out 2281 | no | yes | yes | no |
| `production` | `bounded-range-natural` | pass | pass | pass | 7 | 7 | 59.76 | in 227897 / cached 216576 / out 3912 | no | no | no | no |
| `production` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 10.88 | in 46291 / cached 40704 / out 472 | no | no | no | no |
| `production` | `invalid-input` | pass | pass | pass | 1 | 2 | 9.09 | in 55428 / cached 40704 / out 383 | no | yes | yes | no |
| `generated-client` | `add-two` | pass | pass | pass | 12 | 7 | 64.03 | in 345014 / cached 327296 / out 4284 | no | yes | yes | no |
| `generated-client` | `repeat-add` | pass | pass | pass | 11 | 8 | 43.48 | in 281781 / cached 266112 / out 2317 | no | yes | yes | no |
| `generated-client` | `update-existing` | pass | pass | pass | 13 | 7 | 59.34 | in 391338 / cached 367360 / out 4400 | no | yes | yes | no |
| `generated-client` | `bounded-range` | pass | pass | pass | 8 | 6 | 38.44 | in 151783 / cached 143104 / out 1478 | no | yes | no | no |
| `generated-client` | `bounded-range-natural` | pass | pass | pass | 8 | 6 | 40.33 | in 162033 / cached 149248 / out 2048 | yes | yes | no | no |
| `generated-client` | `ambiguous-short-date` | pass | pass | pass | 1 | 2 | 7.87 | in 46052 / cached 40704 / out 309 | no | no | no | no |
| `generated-client` | `invalid-input` | pass | pass | pass | 5 | 3 | 13.47 | in 71800 / cached 63616 / out 934 | no | yes | yes | no |
| `cli` | `add-two` | pass | pass | pass | 6 | 5 | 33.44 | in 140881 / cached 134400 / out 1127 | no | no | no | no |
| `cli` | `repeat-add` | pass | pass | pass | 4 | 6 | 335.44 | in 140547 / cached 114432 / out 1055 | no | no | no | no |
| `cli` | `update-existing` | pass | pass | pass | 2 | 4 | 23.08 | in 115748 / cached 109952 / out 776 | no | no | no | no |
| `cli` | `bounded-range` | pass | pass | pass | 3 | 4 | 22.04 | in 139180 / cached 133376 / out 655 | no | no | no | no |
| `cli` | `bounded-range-natural` | pass | pass | pass | 2 | 5 | 23.54 | in 115253 / cached 109440 / out 566 | no | no | no | no |
| `cli` | `ambiguous-short-date` | pass | pass | pass | 4 | 3 | 9.74 | in 71105 / cached 63104 / out 759 | no | yes | no | no |
| `cli` | `invalid-input` | pass | pass | pass | 1 | 2 | 8.96 | in 45136 / cached 40704 / out 412 | no | no | no | no |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-2026-04-17.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- |
| `production` | `add-two` | same_pass | -3 | +0 | +13.13 | -3851 | improved_to_no | same_no | same_no | same_no |
| `production` | `repeat-add` | same_pass | -5 | -1 | -0.02 | -6722 | same_no | same_no | same_no | same_no |
| `production` | `update-existing` | same_pass | -9 | +0 | +10.45 | -9259 | improved_to_no | same_no | same_no | same_no |
| `production` | `bounded-range` | same_pass | -5 | -1 | +8.10 | -502 | same_no | regressed_to_yes | regressed_to_yes | improved_to_no |
| `production` | `bounded-range-natural` | same_pass | -1 | +1 | +23.81 | -7081 | improved_to_no | same_no | same_no | same_no |
| `production` | `ambiguous-short-date` | same_pass | +0 | +0 | +0.18 | -10242 | improved_to_no | same_no | same_no | same_no |
| `production` | `invalid-input` | same_pass | -6 | -2 | -10.13 | -422 | same_no | regressed_to_yes | regressed_to_yes | same_no |
| `generated-client` | `add-two` | same_pass | -3 | +0 | +14.32 | -2295 | improved_to_no | regressed_to_yes | regressed_to_yes | same_no |
| `generated-client` | `repeat-add` | same_pass | +0 | +1 | -0.21 | -2566 | improved_to_no | regressed_to_yes | regressed_to_yes | same_no |
| `generated-client` | `update-existing` | same_pass | -1 | -2 | -12.51 | -1264 | improved_to_no | regressed_to_yes | regressed_to_yes | same_no |
| `generated-client` | `bounded-range` | same_pass | -5 | +0 | -12.40 | -10471 | improved_to_no | regressed_to_yes | same_no | same_no |
| `generated-client` | `bounded-range-natural` | same_pass | +3 | +1 | +6.89 | +1705 | regressed_to_yes | regressed_to_yes | same_no | same_no |
| `generated-client` | `ambiguous-short-date` | same_pass | +0 | +0 | -0.66 | +389 | same_no | same_no | same_no | same_no |
| `generated-client` | `invalid-input` | same_pass | +4 | +1 | +6.00 | +3284 | same_no | regressed_to_yes | regressed_to_yes | same_no |
| `cli` | `add-two` | same_pass | -2 | -1 | +2.87 | -2065 | same_no | same_no | same_no | same_no |
| `cli` | `repeat-add` | same_pass | +2 | +2 | +312.95 | +20656 | same_no | same_no | same_no | same_no |
| `cli` | `update-existing` | same_pass | -1 | +0 | -3.91 | +200 | same_no | same_no | same_no | same_no |
| `cli` | `bounded-range` | same_pass | +1 | +0 | +0.12 | +312 | same_no | same_no | same_no | same_no |
| `cli` | `bounded-range-natural` | same_pass | -3 | +0 | -6.84 | -1615 | improved_to_no | same_no | same_no | same_no |
| `cli` | `ambiguous-short-date` | same_pass | +3 | +1 | +0.51 | +3232 | same_no | regressed_to_yes | same_no | same_no |
| `cli` | `invalid-input` | same_pass | +0 | +0 | -2.44 | -266 | same_no | same_no | same_no | same_no |

## Metric Notes

- Production generated paths surfaced from broad repo searches in bounded-range, invalid-input in the 2026-04-17-production-hardening run; this is tracked separately from direct generated-file inspection.
- Production broad repo search remained in bounded-range, invalid-input in the 2026-04-17-production-hardening run.

## Production Stop-Loss

- Triggered: `yes`
- Recommendation: `pivot_to_cli_for_agent_operations`
- Trigger: production update-existing used 14 tools, above threshold 12
- Trigger: production used more than 2x CLI tools in bounded-range, bounded-range-natural, update-existing

## Metric Evidence

- `production/bounded-range` broad repo search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth . | sed -n '1,120p'"`.
- `production/bounded-range` generated path from broad search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth . | sed -n '1,120p'"`.
- `production/invalid-input` broad repo search: `/bin/zsh -lc "pwd && bd prime && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && rg -n \"weight|stone|OpenHealth|entry\" -S . --hidden --glob '"'!**/.git/**'"'"`.
- `production/invalid-input` generated path from broad search: `/bin/zsh -lc "pwd && bd prime && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && rg -n \"weight|stone|OpenHealth|entry\" -S . --hidden --glob '"'!**/.git/**'"'"`.
- `generated-client/add-two` broad repo search: `/bin/zsh -lc "rg --files -g '"'!*node_modules*'"' -g '"'!*.git*'"' | sed -n '1,200p'"`.
- `generated-client/add-two` generated path from broad search: `/bin/zsh -lc "rg --files -g '"'!*node_modules*'"' -g '"'!*.git*'"' | sed -n '1,200p'"`.
- `generated-client/repeat-add` broad repo search: `/bin/zsh -lc "pwd && rg -n \"openhealth|weight|weights|client\" .agents/skills/openhealth/SKILL.md . -g '"'!**/.git/**'"'"`.
- `generated-client/repeat-add` generated path from broad search: `/bin/zsh -lc "pwd && rg -n \"openhealth|weight|weights|client\" .agents/skills/openhealth/SKILL.md . -g '"'!**/.git/**'"'"`.
- `generated-client/update-existing` broad repo search: `/bin/zsh -lc "pwd && bd prime && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && rg -n \"weight|weigh|body weight|03/29/2026|2026-03-29\" -S ."`.
- `generated-client/update-existing` generated path from broad search: `/bin/zsh -lc "pwd && bd prime && sed -n '1,220p' .agents/skills/openhealth/SKILL.md && rg -n \"weight|weigh|body weight|03/29/2026|2026-03-29\" -S ."`.
- `generated-client/bounded-range` broad repo search: `/bin/zsh -lc 'rg -n "OPENHEALTH_(DATA_DIR|DATABASE_PATH)|XDG_DATA_HOME|openhealth.db|LocalConfig" -S .'`.
- `generated-client/bounded-range-natural` direct generated-file: `/bin/zsh -lc "rg -n \"type ListHealthWeightParams|ListHealthWeightWithResponse|CreateHealthWeight\" client internal -g'*.go'"`.
- `generated-client/bounded-range-natural` broad repo search: `/bin/zsh -lc "rg -n \"OPENHEALTH_(DATA_DIR|DATABASE_PATH)|openhealth\\.db|LocalConfig|Paths\\.DatabasePath\" ."`.
- `generated-client/invalid-input` broad repo search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth . | sed 's#"'^./##'"'"`.
- `generated-client/invalid-input` generated path from broad search: `/bin/zsh -lc "pwd && rg --files .agents/skills/openhealth . | sed 's#"'^./##'"'"`.
- `cli/ambiguous-short-date` broad repo search: `/bin/zsh -lc "bd prime && pwd && rg --files -g 'SKILL.md' -g 'AGENTS.md' -g '.agents/**' -g 'README*' -g 'openhealth*' -g '*weight*' -g '*weights*' ."`.

## Scenario Notes

- `production/add-two`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/add-two/events.jsonl`.
- `production/repeat-add`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/production/repeat-add/events.jsonl`.
- `production/update-existing`: expected one updated row; observed [2026-03-29 151.6 lb] Raw event reference: `<run-root>/production/update-existing/events.jsonl`.
- `production/bounded-range`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/bounded-range/events.jsonl`.
- `production/bounded-range-natural`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/production/bounded-range-natural/events.jsonl`.
- `production/ambiguous-short-date`: expected no write and a year clarification; observed [] Raw event reference: `<run-root>/production/ambiguous-short-date/events.jsonl`.
- `production/invalid-input`: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/production/invalid-input/events.jsonl`.
- `generated-client/add-two`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/generated-client/add-two/events.jsonl`.
- `generated-client/repeat-add`: expected exactly two newest-first rows; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] Raw event reference: `<run-root>/generated-client/repeat-add/events.jsonl`.
- `generated-client/update-existing`: expected one updated row; observed [2026-03-29 151.6 lb] Raw event reference: `<run-root>/generated-client/update-existing/events.jsonl`.
- `generated-client/bounded-range`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/generated-client/bounded-range/events.jsonl`.
- `generated-client/bounded-range-natural`: expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed [2026-03-30 151.6 lb, 2026-03-29 152.2 lb, 2026-03-28 153.0 lb] Raw event reference: `<run-root>/generated-client/bounded-range-natural/events.jsonl`.
- `generated-client/ambiguous-short-date`: expected no write and a year clarification; observed [] Raw event reference: `<run-root>/generated-client/ambiguous-short-date/events.jsonl`.
- `generated-client/invalid-input`: expected no write and an invalid input rejection; observed [] Raw event reference: `<run-root>/generated-client/invalid-input/events.jsonl`.
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
