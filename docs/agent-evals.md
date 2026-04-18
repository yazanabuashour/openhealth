# Agent Evaluation Protocol

OpenHealth agent evals measure the same production skill a real agent receives.
Do not add hidden evaluator-only instructions to improve a result; if an
instruction is needed, put it in the production skill first.

## Active Surfaces

- `production`: the installed `skills/openhealth` AgentOps skill.
- `cli`: a CLI baseline skill under `docs/agent-eval-assets/variants/cli`.

The generated client and local SDK remain supported runtime/developer APIs, but
they are not active agent-facing eval variants.

## Scenario Coverage

The `oh-5yr` harness covers routine local user-data tasks:

- weight add/reapply/correction, latest/history/range listing, and invalid input
  rejection
- blood-pressure record, latest/history/range listing, and invalid input
  rejection
- mixed weight and blood-pressure requests in one user task
- true multi-turn requests that require clarification or conversational context

Every scenario uses a fresh ephemeral agent session, an isolated copied repo, a
fresh local database path, and reduced JSON/Markdown artifacts. Raw event logs
are not committed; reduced reports refer to them with `<run-root>` placeholders.
The copied repo intentionally omits root `AGENTS.md`, stale `.agents` content,
and eval/report directories before installing the selected variant skill, so the
production and CLI surfaces do not contaminate each other.

Single-turn scenarios use `codex exec --ephemeral`. Multi-turn scenarios use one
persisted eval session per variant/scenario: the first turn creates a session in
the throwaway run directory context, and later turns use `codex exec resume`
with explicit writable roots for the scenario run directory and shared Go cache.
Per-turn raw logs live under `<run-root>/<variant>/<scenario>/turn-N/`.

The harness runs independent variant/scenario jobs with `--parallel 4` by
default. Use `--parallel 1` when serial execution is needed for debugging or
manual log comparison.

The harness defaults to `--cache-mode shared`, which prewarms one shared Go
module/build cache under `<run-root>/shared-cache` while keeping databases,
temporary directories, repo copies, and raw logs isolated per job. The eval
database lives inside each copied repo at
`<run-root>/<variant>/<scenario>/repo/openhealth.db` so resumed multi-turn
sessions can use the same local path as single-turn runs. Use
`--cache-mode isolated` for apples-to-apples comparison with older reports.

## Metrics

Reports include:

- database verification and assistant-answer verification
- configured harness parallelism and elapsed harness wall time
- cache mode, cache prewarm time, effective parallel speedup, and parallel
  efficiency
- per-job phase timing totals for setup, cache warm, agent run, metrics parsing,
  and verification
- per-turn metrics and raw log references for multi-turn scenarios
- tool calls, assistant calls, wall time, non-cache input tokens, and output
  tokens
- direct generated-file inspection
- generated paths surfaced from broad search
- broad repo search
- Go module-cache inspection
- OpenHealth CLI usage
- direct SQLite access

CLI usage is counted only for executed CLI invocations, not for searches or
documentation reads that merely contain CLI command strings.

## Comparison Policy

Production AgentOps beats CLI only when:

- production passes every selected scenario
- production has no direct generated-file inspection, module-cache inspection,
  direct SQLite access, or CLI usage
- production has no routine broad repo search
- rule-covered invalid-input scenarios are final-answer-only: no tools, no
  command executions, and at most one assistant answer
- production total tools are less than or equal to CLI total tools
- production ties or beats CLI tools in at least 80% of comparable scenarios
- no routine production scenario exceeds CLI by more than one tool
- production has lower non-cached input tokens than CLI in a strict majority of
  comparable scenarios with exposed usage
- production total non-cached input tokens are less than or equal to CLI total
  non-cached input tokens; missing usage on either side fails token comparison

The current weight recommendation is documented in
`docs/agent-eval-results/oh-5yr-agentops-production-expanded.md`. The current
combined weight and blood-pressure result is documented in
`docs/agent-eval-results/oh-5yr-agentops-blood-pressure-expanded.md`.
The current AgentOps runner-overhead follow-up result is documented in
`docs/agent-eval-results/oh-5yr-oh-967-final-r2.md`.
The current eval-maturity and throughput follow-up result is documented in
`docs/agent-eval-results/oh-5yr-maturity-throughput-final.md`.
Historical iteration artifacts are archived under
`docs/agent-eval-results/archive/`.
