- For all committed docs, reports, and artifact references, use repo-relative paths or neutral repo-relative placeholders. Never use machine-absolute filesystem paths.

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

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds, unless the user has explicitly requested a manual review hold after quality gates and issue update.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Optional manual review hold** - Only if the user explicitly requested manual review, stop here in `awaiting manual review` state and do not push yet
5. **PUSH TO REMOTE** - This is MANDATORY once review is approved or if no manual review hold was requested:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
6. **Clean up** - Clear stashes, prune remote branches
7. **Verify** - All changes committed AND pushed, or explicitly paused in `awaiting manual review`
8. **Hand off** - Provide context for next session

**POST-REVIEW RESUME RULES:**
- If review finds issues, address them, rerun the relevant quality gates, update issue status again, and return to the manual review hold if more review was requested
- If review approves, resume at the push step without reopening earlier workflow steps

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds; a session may pause before push only if the user explicitly requested a manual review hold after quality gates and issue update
- If paused for manual review, report the repo state as `awaiting manual review`
- NEVER stop before pushing except for that explicit manual-review hold - that leaves work stranded locally
- NEVER say "ready to push when you are" after review approval - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
