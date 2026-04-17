# Agent Evaluation Plan

This project evaluates agent behavior against the same surface a real OpenHealth
agent receives. The production eval must use only the installed
`skills/openhealth` payload and a fresh session. Do not add hidden evaluator
instructions that tell the agent which API to call unless those instructions are
also present in the production skill.

## Primary Production Eval

- Start from a fresh agent session with the production `openhealth` skill.
- Provide natural user prompts such as `03/29 152.2 lbs and 03/30 151.6`.
- Use the normal local Go/tool environment and default OpenHealth data path.
- Judge success by final database state, newest-first verification, duplicate
  behavior, tool calls, assistant calls, wall time, non-cache input tokens, and
  whether the agent read generated files or the Go module cache.
- The expected production path is `client.OpenLocal(...)` plus the ergonomic
  weight helpers on `LocalClient`.
- Production agents should use the routine weight fast path in
  `skills/openhealth/references/weights.md` before searching the repository.
  Broad repo searches, generated-file inspection, and module-cache inspection
  are tracked as separate metrics.

## Isolated Variants

Keep comparison variants outside the production skill so the real skill stays
opinionated and narrow.

- Baseline A: current or archived generated-client skill surface.
- Variant B: production code-first SDK skill surface.
- Variant C: CLI-oriented skill payload that exercises
  `go run ./cmd/openhealth weight add/list`.

Each variant should have its own skill payload or harness instructions. Do not
combine generated-client, SDK, and CLI recipes in the same production skill.

## Core Scenarios

- Add two weights from a natural-language prompt and verify newest-first output.
- Repeat the same add request and assert no duplicate manual entry is created.
- Add the same date and unit with a different value and assert the existing entry
  is updated through the idempotent path.
- List a bounded date range.
- List the same bounded date range with natural date wording.
- Reject or clarify ambiguous short dates when the year cannot be inferred.
- Reject invalid unit and invalid value inputs.

## Iteration Reports

When a previous reduced JSON report is available, include a comparison section
that calls out pass/fail deltas, tool-call deltas, assistant-call deltas, wall
time deltas, non-cache token deltas, generated-file inspection changes, and
module-cache inspection changes. Treat correctness regressions as blockers;
treat metric-only movement as review context unless generated-file or
module-cache inspection regresses.

## Production Stop-Loss

Give production two focused hardening cycles before recommending a pivot to the
CLI-oriented agent surface. Pivot after the second cycle if production loses
correctness, directly inspects generated files, inspects the module cache, uses
broad repo search in more than one routine scenario, or remains materially worse
than CLI on core tool counts (`add-two` > 10, `update-existing` > 12,
`bounded-range` > 8, or more than 2x CLI tools in at least three comparable
scenarios). Do not add hidden evaluator-only instructions or CLI recipes to the
production skill to avoid the pivot.

As of the production hardening full-matrix eval, the stop-loss is triggered by
the `update-existing` tool-count threshold and production using more than 2x CLI
tools in three comparable scenarios. Production broad repo search remained in
`bounded-range` and `invalid-input`, but `invalid-input` is a validation
scenario and does not satisfy the more-than-one routine scenario trigger. Keep
production as the SDK/developer integration surface, but prefer the CLI-oriented
agent surface for routine local weight operations unless a later eval clears the
stop-loss triggers without hidden evaluator-only instructions.

## Code-First Pivot Trial

The next trial uses a structured code-first AgentOps facade rather than the raw
SDK helper snippets. The candidate variant is `agentops-code`: it creates a
short temporary Go module outside the repository, imports
`github.com/yazanabuashour/openhealth/agentops`, runs a task-shaped
`WeightTaskRequest`, prints JSON, and answers from that JSON only.

Run exactly three code-first iterations, then stop and report whether the final
iteration beat CLI. The candidate beats CLI only if it passes every scenario,
does not directly inspect generated files, does not inspect the Go module cache,
does not use broad repo search in routine scenarios, uses no more total tools
than CLI, is at or below CLI tools in at least five of seven scenarios, and no
routine scenario exceeds CLI by more than one tool.

Status: completed in `docs/agent-eval-results/oh-5yr-code-first-pivot.md`.
The first code-first iteration proved correctness but failed operational
criteria; the second cleared search/cache/generated-file violations but still
lost on tool count and one assistant-format check; the third passed every
criterion. Final measured verdict: prefer `agentops-code` for the routine local
weight tasks covered by `oh-5yr`, while keeping CLI as a human-facing and
fallback path.
