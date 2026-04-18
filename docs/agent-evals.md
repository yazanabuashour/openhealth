# Agent Evaluation Protocol

OpenHealth agent evals measure the same production skill a real client agent
receives. Do not add hidden evaluator-only instructions to improve a result; if
an instruction is needed, put it in the production skill first.

## Active Surface

- `production`: the installed `skills/openhealth/SKILL.md` AgentOps skill plus
  an installed `openhealth` binary on `PATH`.

The generated client and local SDK remain supported runtime/developer APIs, but
they are not active agent-facing eval variants.

## Scenario Coverage

The `oh-5yr` harness covers routine local user-data tasks:

- weight add/reapply/correction, latest/history/range listing, and invalid input
  rejection
- blood-pressure record/correction, latest/history/range listing, and invalid
  input rejection
- medication record/correction/delete/list and invalid input rejection
- lab record/correction/delete/latest/history/range/analyte listing and invalid
  input rejection
- mixed-domain requests in one user task
- true multi-turn requests that require clarification or conversational context

All scenarios are production-only. They gate correctness and hygiene, including
no broad repo search, generated-file inspection, module-cache inspection, direct
SQLite access, or retired human CLI usage.

Every scenario uses a fresh ephemeral agent session, an isolated copied repo, a
fresh local database path, and reduced JSON/Markdown artifacts. Raw event logs
are not committed; reduced reports refer to them with `<run-root>` placeholders.
The copied repo intentionally omits root `AGENTS.md`, stale `.agents` content,
eval assets/results, and the eval harness before installing the selected
variant skill.

The production skill is copied byte-for-byte to
`.agents/skills/openhealth/SKILL.md`. The harness does not generate an
OpenHealth-specific eval `AGENTS.md` or paste skill content into `AGENTS.md`.

Before each production job, the harness builds `openhealth` into the
job's private `bin/` directory and prepends that directory to `PATH`. This
simulates the client-agent install path without requiring the agent to use
`go run`.

The harness renders model-visible context with `codex debug prompt-input` and
fails preflight unless `openhealth` appears as an available project skill, the
skill path points at `.agents/skills/openhealth/SKILL.md`, the installed skill
bytes match `skills/openhealth/SKILL.md`, and no model-visible `AGENTS.md` block
contains OpenHealth runner commands, JSON shapes, validation rules, or
product-agent behavior.

Single-turn scenarios use `codex exec --ephemeral`. Multi-turn scenarios use one
persisted eval session per variant/scenario: the first turn creates a session in
the throwaway run directory context, and later turns use `codex exec resume`
with explicit writable roots for the scenario run directory and shared Go cache.
Per-turn raw logs live under `<run-root>/<variant>/<scenario>/turn-N/`.

The harness runs independent variant/scenario jobs with `--parallel 4` by
default. Use `--parallel 1` when serial execution is needed for debugging or
manual log comparison.

## Metrics

Reports include:

- database verification and assistant-answer verification
- configured harness parallelism and elapsed harness wall time
- cache mode, cache prewarm time, effective parallel speedup, and parallel
  efficiency
- per-job phase timing totals for setup, agent-app build, cache warm, agent run,
  metrics parsing, and verification
- per-turn metrics and raw log references for multi-turn scenarios
- tool calls, assistant calls, wall time, non-cache input tokens, and output
  tokens
- direct generated-file inspection
- generated paths surfaced from broad search
- broad repo search
- Go module-cache inspection
- retired human OpenHealth CLI usage
- direct SQLite access

Retired human CLI usage is counted only for executed old command shapes such as
`openhealth weight add`, `openhealth weight list`, or `go run ./cmd/openhealth`.
Searches or documentation reads that merely contain CLI command strings are not
counted.

## Production Gate

Production AgentOps is release-ready only when:

- production passes every selected scenario
- production has no direct generated-file inspection, module-cache inspection,
  direct SQLite access, or retired human CLI usage
- production has no routine broad repo search
- rule-covered invalid-input scenarios are final-answer-only: no tools, no
  command executions, and at most one assistant answer
- the eval context preflight confirms the client-agent context is the shipped
  project skill, not hidden evaluator-only instructions

Historical iteration artifacts are archived under
`docs/agent-eval-results/archive/`.
