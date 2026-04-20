# oh-5yr Agent Eval Results

Date: oh-cpe-final

Harness: `codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots`

Model: `gpt-5.4-mini`, reasoning effort `medium`

Codex CLI: `codex-cli 0.121.0`

Parallelism: `1`

Cache mode: `shared`

Cache prewarm seconds: `20.65`

Harness elapsed seconds: `94.62`

Effective parallel speedup: `0.86x`

Parallel efficiency: `0.86`

Reduced JSON artifact: `docs/agent-eval-results/oh-5yr-oh-cpe-final.json`

Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.

## History Isolation

- Status: `passed`
- Single-turn ephemeral runs: `5`.
- Multi-turn persisted sessions: `1` sessions / `2` turns.
- The Codex desktop app was not used for eval prompts.
- New Codex session files referencing `<run-root>`: `1`.

## Results

| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |
| `production` | `non-iso-date-reject` | pass | pass | pass | 0 | 2 | 10.84 | in 22041 / cached 18816 / out 301 | no | no | no | no | no | no |
| `production` | `mixed-invalid-direct-reject` | pass | pass | pass | 0 | 1 | 5.92 | in 22038 / cached 18816 / out 341 | no | no | no | no | no | no |
| `production` | `lab-invalid-slug` | pass | pass | pass | 0 | 1 | 5.86 | in 22040 / cached 18816 / out 211 | no | no | no | no | no | no |
| `production` | `bp-add-two` | pass | pass | pass | 2 | 3 | 15.57 | in 73787 / cached 66176 / out 681 | no | no | no | no | no | no |
| `production` | `mixed-medication-lab` | pass | pass | pass | 4 | 3 | 17.03 | in 101560 / cached 93184 / out 1077 | no | no | no | no | no | no |
| `production` | `mt-mixed-latest-then-correct` | pass | pass | pass | 5 | 5 | 26.21 | in 279165 / cached 262784 / out 1770 | no | no | no | no | no | no |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.17 |
| install_variant | 0.00 |
| build_agent_app | 5.25 |
| warm_cache | 0.00 |
| seed_db | 0.06 |
| agent_run | 81.43 |
| parse_metrics | 0.00 |
| verify | 0.04 |
| total_job_time | 94.61 |

## Comparison

Baseline: `docs/agent-eval-results/oh-5yr-2026-04-19-v0.1.0-final.json`.

| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `production` | `non-iso-date-reject` | same_pass | -1 | +0 | +0.95 | -2769 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `mixed-invalid-direct-reject` | same_pass | -1 | -1 | -1.08 | -2683 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `lab-invalid-slug` | same_pass | -1 | -1 | -2.49 | -2747 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `bp-add-two` | same_pass | -4 | -2 | -9.03 | -1360 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `mixed-medication-lab` | same_pass | -2 | -1 | -5.50 | -6802 | same_no | same_no | same_no | same_no | same_no | same_no |
| `production` | `mt-mixed-latest-then-correct` | same_pass | +0 | -1 | -13.16 | -4697 | same_no | same_no | same_no | same_no | same_no | same_no |

## Production Stop-Loss

- Triggered: `no`
- Recommendation: `ship_openhealth_runner_production`

## Turn Details

- `production/mt-mixed-latest-then-correct` turn 1: exit `0`, tools `3`, assistant calls `3`, wall `13.74`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-1/events.jsonl`.
- `production/mt-mixed-latest-then-correct` turn 2: exit `0`, tools `2`, assistant calls `2`, wall `12.47`, raw `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.

## Scenario Notes

- `production/non-iso-date-reject`: turn 1: expected no write and a strict YYYY-MM-DD date rejection; observed [] Raw event reference: `<run-root>/production/non-iso-date-reject/turn-1/events.jsonl`.
- `production/mixed-invalid-direct-reject`: turn 1: expected no mixed-domain writes and a direct invalid input rejection; observed weights [] and blood pressures [] Raw event reference: `<run-root>/production/mixed-invalid-direct-reject/turn-1/events.jsonl`.
- `production/lab-invalid-slug`: turn 1: expected no write and invalid analyte slug rejection; observed [] Raw event reference: `<run-root>/production/lab-invalid-slug/turn-1/events.jsonl`.
- `production/bp-add-two`: turn 1: expected exactly two newest-first blood-pressure rows; observed [2026-03-30 118/76, 2026-03-29 122/78 pulse 64 note home cuff] Raw event reference: `<run-root>/production/bp-add-two/turn-1/events.jsonl`.
- `production/mixed-medication-lab`: turn 1: expected one medication and one glucose lab; observed medications [Levothyroxine 25 mcg from 2026-01-01] and labs [2026-03-29 (Glucose 89)] Raw event reference: `<run-root>/production/mixed-medication-lab/turn-1/events.jsonl`.
- `production/mt-mixed-latest-then-correct`: turn 1: expected unchanged seed rows and latest mixed-domain answer; observed weights [2026-03-30 151.6 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 118/76, 2026-03-29 122/78 pulse 64]; turn 2: expected latest mixed-domain corrections on 2026-03-30; observed weights [2026-03-30 151.0 lb, 2026-03-29 152.2 lb] and blood pressures [2026-03-30 117/75 pulse 63, 2026-03-29 122/78 pulse 64] Raw event reference: `<run-root>/production/mt-mixed-latest-then-correct/turn-2/events.jsonl`.

## App Server

not used: codex exec --json exposed enough event detail for this run.
