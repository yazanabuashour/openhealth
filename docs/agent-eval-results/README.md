# Agent Eval Results

Current recommendation:

- Use the production AgentOps skill for routine local weight tasks.
- Use the production AgentOps skill for routine local blood-pressure tasks.
- Keep the CLI as human-facing tooling and as the eval baseline. It is not a
  production skill fallback.
- Current blood-pressure expansion verdict: AgentOps is correct and clean enough
  for production agents, but it does not beat the optimized CLI baseline on tool,
  speed, or token efficiency.

Current top-level reports:

- `docs/agent-eval-results/oh-5yr-agentops-production-expanded.md`
- `docs/agent-eval-results/oh-5yr-agentops-blood-pressure-expanded.md`

Historical iteration artifacts live in `docs/agent-eval-results/archive/`.
Those files preserve provenance for earlier SDK, generated-client, CLI, and
AgentOps pivot experiments without making the primary results directory the main
reading path.
