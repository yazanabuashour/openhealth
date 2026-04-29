# OpenHealth Taste Review Backlog

## Status

Planning backlog created after the v0.5.0 production eval and recent OpenClerk
taste-review correction.

This note records a process improvement, not a new public API. It keeps the
existing OpenHealth ADR, POC, eval, decision, and implementation follow-up
workflow, while adding a clearer taste review for cases where OpenHealth is
technically safe but unnecessarily awkward.

## Baseline Lesson

OpenHealth already has a strong evidence process: production evals verify the
shipped runner/skill surface, reduced reports keep raw logs out of the repo,
and release gates check correctness plus hygiene. Keep that process.

The correction is that a workflow can pass safety and capability checks while
still creating taste debt. A technically valid health-data workflow may be too
ceremonial if it requires many runner calls, long latency, exact prompt
choreography, or clarification turns that a normal user would not expect.

The approval boundary matters:

- inspect, read, and list operations are different from durable writes
- writes, corrections, deletes, imports, and unsafe mutation remain explicit
  approval boundaries
- runner-only access, local-first storage, validation, provenance,
  auditability, and no direct SQLite access still apply
- passing eval classification is not proof that the routine UX is good

## Taste Review Lens

Future deferral, reference, promotion, or release decisions should ask one more
question after the safety and capability checks: would a normal user reasonably
expect a simpler OpenHealth surface here?

Useful signals include:

- the workflow passes but needs high tool count, repeated assistant turns, or
  long wall time
- user intent fits the natural scope of an existing runner domain, but the
  current policy makes the agent route through ceremony
- the agent asks for a clarification that protects safety, but the prompt shape
  suggests a better infer, normalize, ask, or reject ladder may exist
- the system preserves data correctly but places naming, note text, dosage
  detail, body-site text, or import context in a non-obvious way
- a success depends on exact prompt choreography that routine users would not
  know

Taste debt does not automatically authorize implementation. It creates audit,
design, or eval backlog unless targeted evidence and a promotion decision name
the exact smoother surface and show that safety, provenance, authority,
local-first operation, runner-only access, and explicit approval boundaries
remain intact.

## Tracker Backlog

The following Beads epics track the revisit work:

- `oh-6wz`: Re-audit OpenHealth intake and validation UX
- `oh-ik1`: Re-audit naming, placement, and capture UX
- `oh-br4`: Re-audit high-touch successful workflows
- `oh-6u0`: Update OpenHealth decision process for taste

These epics are docs and evaluation-design backlog only. They do not authorize
runner actions, schema changes, storage migrations, skill behavior changes,
runner/API changes, or implementation follow-up. Any future implementation
still needs targeted evidence and an explicit promotion decision naming the
exact surface and gates.

## Initial Audit Targets

Re-audit intake and validation UX for strict rejection and clarification
decisions: short dates without a year, year-first slash dates, unsupported
units, invalid values, invalid lab slugs, and mixed-domain invalid writes. The
question is whether the current infer, normalize, ask, or reject ladder is
obvious to routine users while preserving safety.

Re-audit naming, placement, and capture UX around lab analyte slug naming, note
placement across domains, medication dosage text, imaging title and body-site
placement, and body-composition plus weight split decisions. The question is
when OpenHealth should infer, propose, preserve, or ask.

Re-audit high-touch successful workflows where eval rows passed but were
expensive or brittle. Initial candidates from
`docs/agent-eval-results/oh-5yr-2026-04-24-v0.5.0-final.md` include
`mixed-import-file-coverage`, `mt-mixed-latest-then-correct`, `lab-patch`,
`medication-note`, `medication-delete`, `lab-same-day-multiple`, and imaging
record/correct/delete flows.

Update process docs so future eval reports and decision notes separately
record:

- safety pass: the workflow preserved validation, provenance, local runner
  boundaries, auditability, and bypass protections
- capability pass: current primitives can technically express the workflow
- UX quality: the workflow is acceptable for routine use or should be recorded
  as taste debt
