# oh-5yr Code-First Pivot Trial

This report documents the code-first pivot for the `oh-5yr` eval after the
production SDK variant reached the hardening stop-loss. Reduced artifacts:

- Iteration 1: `docs/agent-eval-results/archive/oh-5yr-code-first-iter1.json` and `docs/agent-eval-results/archive/oh-5yr-code-first-iter1.md`
- Iteration 2: `docs/agent-eval-results/archive/oh-5yr-code-first-iter2.json` and `docs/agent-eval-results/archive/oh-5yr-code-first-iter2.md`
- Iteration 3: `docs/agent-eval-results/archive/oh-5yr-code-first-iter3.json` and `docs/agent-eval-results/archive/oh-5yr-code-first-iter3.md`
- Prior production hardening: `docs/agent-eval-results/archive/oh-5yr-2026-04-17-production-hardening.json` and `docs/agent-eval-results/archive/oh-5yr-2026-04-17-production-hardening.md`

Raw logs are not committed. Raw event references in reduced reports use
`<run-root>` placeholders.

## Current Findings Before Pivot

The production SDK path completed the baseline run plus two focused hardening
iterations. The final hardening run was correct across all seven scenarios, but
it did not beat CLI on operational efficiency. Its stop-loss triggered because
`production/update-existing` used 14 tools, above the 12-tool threshold, and
production used more than 2x CLI tools in `bounded-range`,
`bounded-range-natural`, and `update-existing`.

The production run also still showed broad repo search in `bounded-range` and
`invalid-input`, with generated paths surfaced from those broad searches. The
pre-pivot verdict was therefore: keep CLI as the preferred routine local weight
operation path until a different code-first approach clears the same thresholds.

## Research Basis

The pivot used a structured, task-level code facade rather than asking agents to
discover lower-level SDK details. This was based on:

- SWE-agent's Agent-Computer Interface framing: agent performance depends on
  LM-centric command and feedback design, and simple purpose-built interfaces
  can improve agent behavior ([docs](https://swe-agent.com/0.7/background/aci/),
  [paper](https://arxiv.org/abs/2405.15793)).
- Anthropic's tool guidance: implement task-shaped tools, consolidate repeated
  workflows, return high-signal context, and iterate with evals
  ([article](https://www.anthropic.com/engineering/writing-tools-for-agents),
  [tool docs](https://platform.claude.com/docs/en/agents-and-tools/tool-use/define-tools)).
- OpenAI's function and structured output guidance: define typed schemas with
  clear names and descriptions, keep tool surfaces constrained, and use evals to
  choose the structure that works for the use case
  ([function calling](https://developers.openai.com/api/docs/guides/function-calling),
  [structured outputs](https://developers.openai.com/api/docs/guides/structured-outputs),
  [evals](https://developers.openai.com/api/docs/guides/evals)).
- GeoJSON Agents' function-calling vs code-generation benchmark: code generation
  can be more flexible for open-ended tasks while function calls are steadier
  for structured operations, motivating a hybrid code-first runner over a narrow
  structured operation API ([paper](https://arxiv.org/abs/2509.08863)).

The resulting approach was a narrow Go `agentops` package with one public
weight-task entry point and deterministic JSON-friendly output.

## Implementation

The new `agentops` package exposes:

```go
RunWeightTask(ctx context.Context, config client.LocalConfig, request WeightTaskRequest) (WeightTaskResult, error)
```

`WeightTaskRequest` supports task-level upsert, latest/history/range list, and
validation-only rejection flows. Inputs are simple strings and primitives:
strict ISO `YYYY-MM-DD` dates, positive numeric values, and units normalized only
from `lb`, `lbs`, `pound`, or `pounds`. Invalid dates, non-positive values, and
unsupported units return `Rejected: true` with a stable reason and perform no
write.

The new `agentops-code` eval variant tells the agent to create a temporary Go
module outside the repo, import `github.com/yazanabuashour/openhealth/agentops`,
print the JSON result, and answer only from that JSON. The variant forbids CLI
usage, generated-file inspection, module-cache inspection, direct SQLite writes,
and broad repo search.

The harness now includes `agentops-code` as a selectable variant and writes a
`code_first` JSON/Markdown comparison against CLI.

## Iterations

| Iteration | Target | Result |
| --- | --- | --- |
| 1 | First `agentops` package and `agentops-code` skill. | Correctness passed, but it lost to CLI: 72 tools vs 22, only 2 of 7 scenarios at or below CLI, direct generated-file inspection, module-cache inspection, and routine broad search all occurred. |
| 2 | Tighten skill/API ergonomics and inject self-contained `AGENTS.md` guidance for the eval copy. | Search/cache/generated-file violations cleared, but it still lost: 32 tools vs 26, only 3 of 7 scenarios at or below CLI, and `bounded-range-natural` failed assistant verification because the final answer was raw JSON rather than plain bounded rows. |
| 3 | Final hardening: exact offline Go runner, `context.Background()`, and plain newest-first row output from JSON entries. | Beat CLI on every defined criterion: all scenarios passed, no direct generated-file inspection, no module-cache inspection, no broad repo search in routine scenarios, 7 tools vs CLI's 28, and 7 of 7 scenarios at or below CLI tools. |

Iteration 3 scenario tool counts:

| Scenario | `agentops-code` | `cli` | Delta |
| --- | ---: | ---: | ---: |
| `add-two` | 1 | 7 | -6 |
| `repeat-add` | 2 | 6 | -4 |
| `update-existing` | 1 | 8 | -7 |
| `bounded-range` | 1 | 2 | -1 |
| `bounded-range-natural` | 2 | 2 | 0 |
| `ambiguous-short-date` | 0 | 1 | -1 |
| `invalid-input` | 0 | 2 | -2 |

## Final Verdict

The structured code-first AgentOps facade prevailed over CLI in the final
iteration for this eval scope. The final recommendation from the reduced report
is `prefer_agentops_code_for_routine_weight_operations`.

This supersedes the pre-pivot CLI preference only for the routine local weight
tasks covered by `oh-5yr`. CLI should remain available as a human-facing and
fallback path. Before generalizing this result, expand coverage beyond these
seven scenarios and run repeated samples, because the current evidence is one
three-iteration trial with inferred tool metrics from Codex JSON events.
