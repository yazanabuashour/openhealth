# Agent Eval Results

Current recommendation:

- Use the production OpenHealth runner/skill for routine local weight tasks.
- Use the production OpenHealth runner/skill for routine local blood-pressure
  tasks.
- Keep the CLI as human-facing tooling and as the eval baseline. It is not a
  production skill fallback.
- Current release gate: the production runner/skill passed all 43 production
  scenarios in `docs/agent-eval-results/oh-5yr-2026-04-19-v0.1.0-final.md`.
- Current maturity/throughput verdict: in the CLI comparison report, the runner
  matched or improved correctness while using fewer tools, fewer non-cached
  input tokens, and less wall time than CLI.

Current top-level reports:

- `docs/agent-eval-results/oh-5yr-2026-04-19-v0.1.0-final.md`
- `docs/agent-eval-results/oh-5yr-2026-04-19.md`
- `docs/agent-eval-results/oh-5yr-maturity-throughput-final.md`
- `docs/agent-eval-results/oh-5yr-2026-04-18.md`

Historical iteration artifacts live in `docs/agent-eval-results/archive/`.
Those files preserve provenance for earlier SDK, generated-client, CLI, and
runner pivot experiments without making the primary results directory the main
reading path. Historical filenames and archived report contents are left as-is
when renaming would corrupt provenance.
