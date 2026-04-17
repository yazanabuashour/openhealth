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
