# oh-5yr AgentOps Blood-Pressure Expansion

## Verdict

`agentops` now covers both routine weight and blood-pressure tasks, and the
optimized production skill was correct in all final samples with clean hygiene.
It did not beat the optimized CLI baseline on the full efficiency threshold:
production used more tools, more wall time, and more non-cached tokens.

Practical recommendation:

- Use AgentOps as the production agent-facing interface for supported weight and
  blood-pressure workflows because it is correct and avoids CLI fallback,
  generated-file inspection, module-cache inspection, broad search, and direct
  SQLite access.
- Keep CLI as human/baseline tooling and the current efficiency winner for
  simple local tasks.
- Do not claim the production AgentOps skill beats CLI under the current
  `oh-5yr` comparison criteria.

## Artifacts

Retained reduced reports:

- Baseline diagnostic sample:
  `docs/agent-eval-results/archive/oh-5yr-agentops-bp-baseline-r1.json`
  and `docs/agent-eval-results/archive/oh-5yr-agentops-bp-baseline-r1.md`
- Optimized final samples:
  `docs/agent-eval-results/archive/oh-5yr-agentops-bp-final-r1.json`
  and `docs/agent-eval-results/archive/oh-5yr-agentops-bp-final-r1.md`
- `docs/agent-eval-results/archive/oh-5yr-agentops-bp-final-r2.json`
  and `docs/agent-eval-results/archive/oh-5yr-agentops-bp-final-r2.md`
- `docs/agent-eval-results/archive/oh-5yr-agentops-bp-final-r3.json`
  and `docs/agent-eval-results/archive/oh-5yr-agentops-bp-final-r3.md`

Raw logs remain outside the repo and are referenced by reduced reports with
`<run-root>` placeholders.

## Protocol Note

The first attempted baseline was aborted before a committed report because the
CLI variant was contaminated by root `AGENTS.md` AgentOps guidance. The harness
now omits root `AGENTS.md`, stale `.agents`, eval docs, reports, and harness
code from throwaway eval copies before installing the selected variant skill.

The retained baseline sample also exposed a verifier false negative for
blood-pressure answers that format a date heading on one line and the reading on
the next line. The verifier was fixed before the optimized final samples. That
baseline sample is retained as diagnostic provenance rather than as a final
decision sample.

## Baseline Diagnostic

| Variant | Correct | Tools | Wall Seconds | Non-Cached Input | Output Tokens | Broad Searches |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `production` | 16/17* | 58 | 402.3 | 144,862 | 19,780 | 2 |
| `cli` | 17/17 | 108 | 487.0 | 188,443 | 20,741 | 9 |

`*` The production miss was the verifier false negative described above. A
focused smoke run after the verifier fix passed the affected case and removed
the two routine broad searches that appeared in `update-existing` and
`bp-add-two`.

## Final Aggregate

Three optimized final samples ran the full combined suite: 10 weight scenarios
and 7 blood-pressure scenarios for each variant.

| Variant | Correct | Avg Tools/Run | Avg Wall Seconds/Run | Avg Non-Cached Input/Run | Avg Output/Run | Broad Searches | CLI Usage | Direct SQLite |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `production` | 51/51 | 49.0 | 399.6 | 125,386 | 17,912 | 0 | 0 | 0 |
| `cli` | 50/51 | 41.3 | 314.1 | 92,979 | 10,501 | 6 | 41 | 0 |

Production passed every final scenario and had zero direct generated-file
inspection, zero module-cache inspection, zero broad repo search, zero CLI use,
and zero direct SQLite access.

CLI was faster and more token/tool efficient, but had one correctness miss in
`bp-add-two` in final sample 3: it duplicated the 2026-03-29 blood-pressure
reading and verification observed three rows instead of two.

## Final Criteria

| Criterion | Result |
| --- | --- |
| Production passes every scenario | Pass |
| No direct generated-file inspection | Pass |
| No module-cache inspection | Pass |
| No routine broad repo search | Pass |
| No production CLI usage | Pass |
| No direct SQLite access | Pass |
| Production total tools <= CLI total tools | Fail: 147 vs 124 |
| Production at/below CLI tools in at least 80% of scenarios | Fail: 5/17 |
| No routine production scenario exceeds CLI by more than one tool | Pass |

## Interpretation

AgentOps is now the safer production interface: it made the supported task space
explicit, rejected invalid inputs before writes, avoided forbidden surfaces, and
was more correct than CLI across the final samples.

CLI remains the efficiency baseline because simple operations often take one
skill read plus one command, while AgentOps usually takes one skill/reference
read plus a temporary Go runner. That extra runner step is realistic for a
code-first skill, but it costs wall time and tokens.

The eval is representative of a skill-driven agent workflow: each scenario uses
an ephemeral agent, an isolated copied repo, a configured local database path,
and the installed variant skill. It is not a benchmark for direct library calls
inside an application that already imports `agentops`; in that setting AgentOps
would avoid the temporary-module setup cost.
