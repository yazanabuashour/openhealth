- For all committed docs, reports, and artifact references, use repo-relative paths or neutral repo-relative placeholders. Never use machine-absolute filesystem paths.
- Do work on the current branch. Do not create or switch to another branch unless explicitly instructed.

## OpenHealth User Data Requests

For direct local OpenHealth weight, blood-pressure, medication, or lab requests,
act as a product data agent, not a repo maintainer. Do not run `bd prime`,
inspect `.agents`, source/generated files, the Go module cache, or SQLite, or
search the repo before the first runner call.

Reject final-answer-only, with exactly one assistant answer and no tools or DB
check, for ambiguous short dates with no year, year-first slash dates like
`2026/03/31`, non-positive values, unsupported units, unsupported lab analyte
slugs, unsupported medication status, empty optional text fields, or medication
end dates before start dates. Do not first announce skill use or process.
`03/29/2026` may become `2026-03-29`.

For valid tasks, pipe JSON to `go run ./cmd/openhealth-agentops weight`,
`go run ./cmd/openhealth-agentops blood-pressure`,
`go run ./cmd/openhealth-agentops medications`, or
`go run ./cmd/openhealth-agentops labs`. Use one call per domain for mixed
requests and answer from JSON only; `entries` are newest-first. Use history with
`limit:2` for "two most recent"; latest returns one row.

Every request JSON must include `action`. Exact one-line shapes:
`{"action":"upsert_weights","weights":[{"date":"2026-03-29","value":152.2,"unit":"lb"}]}`;
`{"action":"list_weights","list_mode":"latest"}`;
`{"action":"list_weights","list_mode":"history","limit":2}`;
`{"action":"list_weights","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}`;
`{"action":"record_blood_pressure","readings":[{"date":"2026-03-29","systolic":122,"diastolic":78,"pulse":64}]}`;
`{"action":"correct_blood_pressure","readings":[{"date":"2026-03-29","systolic":121,"diastolic":77,"pulse":63}]}`;
`{"action":"list_blood_pressure","list_mode":"latest"}`;
`{"action":"list_blood_pressure","list_mode":"history","limit":2}`;
`{"action":"list_blood_pressure","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}`;
`{"action":"record_medications","medications":[{"name":"Levothyroxine","dosage_text":"25 mcg","start_date":"2026-01-01"}]}`;
`{"action":"correct_medication","target":{"name":"Levothyroxine","start_date":"2026-01-01"},"medication":{"name":"Levothyroxine","dosage_text":"50 mcg","start_date":"2026-01-01","end_date":"2026-04-01"}}`;
`{"action":"delete_medication","target":{"name":"Levothyroxine","start_date":"2026-01-01"}}`;
`{"action":"list_medications","status":"active"}`;
`{"action":"list_medications","status":"all"}`;
`{"action":"record_labs","collections":[{"date":"2026-03-29","panels":[{"panel_name":"Metabolic","results":[{"test_name":"Glucose","canonical_slug":"glucose","value_text":"89","value_numeric":89,"units":"mg/dL","range_text":"70-99"}]}]}]}`;
`{"action":"correct_labs","target":{"date":"2026-03-29"},"collection":{"date":"2026-03-29","panels":[{"panel_name":"Thyroid","results":[{"test_name":"TSH","canonical_slug":"tsh","value_text":"3.1","value_numeric":3.1,"units":"uIU/mL"}]}]}}`;
`{"action":"delete_labs","target":{"date":"2026-03-29"}}`;
`{"action":"list_labs","list_mode":"latest"}`;
`{"action":"list_labs","list_mode":"history","limit":2}`;
`{"action":"list_labs","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}`;
`{"action":"list_labs","list_mode":"latest","analyte_slug":"glucose"}`.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for maintainer issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- If you are acting as a maintainer or local coding agent, use `bd` for task tracking instead of ad hoc markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete steps 1-5 below, then stop for manual review before running `git add`, `git commit`, `bd dolt push`, or `git push`. The workflow is paused for manual review at step 5 with uncommitted local changes, and the work session is NOT complete until steps 6-10 are finished after review approval and `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Prepare manual review** - Run `git status`, summarize changed files and quality gates, confirm no commit or push has been performed, and leave files uncommitted for manual review
5. **Manual review** - Stop here by default with uncommitted local changes, report that the workflow is paused for manual review, and wait for explicit instruction to complete the remaining steps
6. **Commit approved changes** - After explicit review approval, stage the intended files and create a local commit
7. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
8. **Clean up** - Clear stashes, prune remote branches
9. **Verify** - All changes committed AND pushed
10. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- The workflow pauses for manual review after step 5 with uncommitted local changes, and the work session is NOT complete until `git push` succeeds
- Do NOT run `git add`, `git commit`, `bd dolt push`, or `git push` before manual review unless explicitly instructed
- Do NOT continue past `Manual review` unless explicitly instructed to complete the remaining workflow steps
- Once instructed to continue after review, stage, commit, pull/rebase, run `bd dolt push`, and `git push`; do NOT stop again with local-only changes
- NEVER say "ready to push when you are" after review approval - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
