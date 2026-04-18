# oh-5yr AgentOps Production Expanded Evaluation

This report documents `oh-874`: promoting the top-performing structured
AgentOps weight facade into the production OpenHealth skill and rerunning
expanded production-vs-CLI samples.

Reduced artifacts:

- Run 1: `docs/agent-eval-results/archive/oh-5yr-agentops-production-expanded-r1.json` and `docs/agent-eval-results/archive/oh-5yr-agentops-production-expanded-r1.md`
- Run 2: `docs/agent-eval-results/archive/oh-5yr-agentops-production-expanded-r2.json` and `docs/agent-eval-results/archive/oh-5yr-agentops-production-expanded-r2.md`
- Run 3: `docs/agent-eval-results/archive/oh-5yr-agentops-production-expanded-r3.json` and `docs/agent-eval-results/archive/oh-5yr-agentops-production-expanded-r3.md`
- Prior pivot report: `docs/agent-eval-results/archive/oh-5yr-code-first-pivot.md`

Raw logs are not committed. Raw event references in reduced reports use
`<run-root>` placeholders.

## Scope

The production `skills/openhealth` payload now directs routine local weight
add, reapply, correction, latest, history, bounded-range, and validation
requests through:

```go
agentops.RunWeightTask(context.Background(), client.LocalConfig{}, request)
```

This report predates the blood-pressure expansion. Its conclusion remains
narrowly scoped to weight. The combined weight and blood-pressure result is in
`docs/agent-eval-results/oh-5yr-agentops-blood-pressure-expanded.md`.

## Protocol

Each sample compared `production` directly against `cli`:

```bash
go run ./scripts/agent-eval/oh5yr run \
  --date agentops-production-expanded-rN \
  --variant production,cli \
  --candidate production \
  --compare-to <previous-report>.json \
  --run-root <run-root>
```

The expanded matrix used ten scenarios:

- `add-two`
- `repeat-add`
- `update-existing`
- `bounded-range`
- `bounded-range-natural`
- `latest-only`
- `history-limit-two`
- `ambiguous-short-date`
- `invalid-input`
- `non-iso-date-reject`

The candidate had to pass every scenario, avoid generated-file inspection,
module-cache inspection, routine broad repo search, OpenHealth CLI usage, and
direct SQLite access, use no more total tools than CLI, tie or beat CLI tools in
at least 8 of 10 scenarios, and avoid exceeding CLI by more than one tool in any
routine scenario.

## Implementation Changes

- Updated `skills/openhealth/SKILL.md` and
  `skills/openhealth/references/weights.md` so the production skill defaults to
  a one-command AgentOps temp Go runner for routine weight tasks.
- Added repo-level user-data guidance in `AGENTS.md` so fresh production eval
  sessions see the same instruction without hidden evaluator-only prompts.
- Kept the CLI as the baseline and fallback, but removed it from the production
  routine weight path.
- Extended `scripts/agent-eval/oh5yr` with the `--candidate` comparison mode,
  extra scenarios, and hygiene metrics for CLI usage and direct SQLite access.
- Hardened generated-file and direct-SQLite metric detection to avoid counting
  references from skill text or schema/table-name searches as direct misuse.

## Results

| Run | Beats CLI | Production Tools | CLI Tools | Production Seconds | CLI Seconds | Production Non-Cache Input | CLI Non-Cache Input | Production Output | CLI Output |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `r1` | yes | 17 | 92 | 190.92 | 641.69 | 65,808 | 164,623 | 8,836 | 56,782 |
| `r2` | yes | 15 | 74 | 175.62 | 460.30 | 62,798 | 122,368 | 8,336 | 37,806 |
| `r3` | yes | 16 | 74 | 163.66 | 466.64 | 48,995 | 107,578 | 7,498 | 37,042 |

Aggregate:

| Variant | Scenarios | Passed | DB Pass | Assistant Pass | Tools | Assistant Calls | Wall Seconds | Non-Cache Input | Output Tokens | Broad Search | Generated Files | Generated From Broad | Module Cache | CLI Used | Direct SQLite |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `production` | 30 | 30 | 30 | 30 | 48 | 92 | 530.20 | 177,601 | 24,670 | 0 | 0 | 0 | 0 | 0 | 0 |
| `cli` | 30 | 30 | 30 | 30 | 240 | 180 | 1,568.63 | 394,569 | 131,630 | 4 | 0 | 1 | 0 | 4 | 0 |

Scenario tool totals across all three runs:

| Scenario | Production Tools | CLI Tools | Production Wins/Ties |
| --- | ---: | ---: | ---: |
| `add-two` | 7 | 32 | 3/3 |
| `repeat-add` | 6 | 37 | 3/3 |
| `update-existing` | 6 | 30 | 3/3 |
| `bounded-range` | 6 | 24 | 3/3 |
| `bounded-range-natural` | 6 | 42 | 3/3 |
| `latest-only` | 6 | 30 | 3/3 |
| `history-limit-two` | 6 | 40 | 3/3 |
| `ambiguous-short-date` | 0 | 0 | 3/3 |
| `invalid-input` | 3 | 3 | 3/3 |
| `non-iso-date-reject` | 2 | 2 | 2/3 |

## Efficiency

Production used 48 tools versus CLI's 240 across the 30 scenario runs, an 80%
reduction. Mean tool use was 1.6 tools per scenario for production versus 8.0
for CLI.

Production accumulated 530.20 scenario wall seconds versus CLI's 1,568.63, about
66% lower. Mean scenario wall time was 17.67 seconds for production versus 52.29
seconds for CLI.

Production used 177,601 non-cache input tokens versus CLI's 394,569, about 55%
lower. Output tokens dropped from 131,630 to 24,670, about 81% lower. The large
output reduction mostly comes from the production path avoiding repeated
repository exploration and answering directly from the AgentOps JSON payload.

## Correctness And Hygiene

Both variants were correct in all 30 sampled runs: database verification passed
30/30, assistant-answer verification passed 30/30, and overall scenario result
passed 30/30 for both production and CLI.

The production AgentOps path also cleared every hygiene criterion in all three
runs: no direct generated-file inspection, no generated path surfaced from broad
search, no Go module cache inspection, no routine broad repo search, no
OpenHealth CLI use, and no direct SQLite access.

CLI remained correct, but it still showed the expected operational noise for an
agent surface that encourages command discovery: four broad repo-search hits,
one generated-path-from-broad-search hit, and four detected OpenHealth CLI
command uses.

## Final Verdict

Prefer the production AgentOps facade for routine local weight operations covered
by this matrix. This supersedes the earlier CLI preference for the measured
weight scope because the production path now matched CLI correctness while using
far fewer tools, less wall time, fewer tokens, and cleaner repository hygiene in
three repeated samples.

Keep the CLI as a human-facing and fallback interface. Do not generalize this
result beyond local weight operations until comparable scenarios and repeated
samples exist for the next domain.
