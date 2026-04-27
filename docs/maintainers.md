# Maintainer Notes

This repository uses **Beads** (`bd`) in embedded mode for maintainer task tracking.

This repository is public and includes a production `openhealth` runner binary,
a single-file OpenHealth skill, a direct-local Go package, and a local SQLite
runtime. Keep maintainer docs honest about the actual supported surface.

Recurring security operations are tracked in
[docs/security-operations.md](security-operations.md). Use that runbook for
dependency review cadence, advisory rehearsal, threat-model refreshes, and
deeper testing expectations.

## Initial Setup

Preferred tool install:

```bash
mise install
```

Alternative:

```bash
brew install beads dolt
```

## Clone Bootstrap

For a fresh maintainer clone or a second machine:

```bash
git clone git@github.com:yazanabuashour/openhealth.git
cd openhealth
bd bootstrap
bd hooks install
```

If role detection warns in a maintainer clone, set:

```bash
git config beads.role maintainer
```

## Sync Between Machines

Push local Beads state before switching machines, then pull on the other machine:

```bash
bd dolt push
bd dolt pull
```

If `bd dolt pull` reports uncommitted Dolt changes, commit them first and retry:

```bash
bd dolt commit
bd dolt pull
```

## Public repo expectations

- Outside contributors must be able to contribute without Beads.
- Policy and workflow files are part of the public contract and should stay reviewable in Git alone.
- Do not document machine-absolute filesystem paths in committed docs.
- Do not assume private infrastructure, deploy secrets, or internal services exist unless they have been added explicitly.

## Repository administration

Current readiness assumptions:

- `main` is the protected default branch.
- Pull requests run only untrusted-safe validation with read-only token scope.
- Pull requests enforce storage codegen drift checks through `go generate ./...` plus `git diff --exit-code`.
- GitHub Releases are created from version tags in the `v0.y.z` form.
- Release publication starts automatically from version tags after release-input verification passes, with narrowly scoped write permissions only on the trusted tag workflow.
- Security reports are expected through GitHub private vulnerability reporting.

Current review enforcement nuance:

- The repository currently has a single maintainer account.
- `main` requires pull requests, status checks, conversation resolution, and one approving review, but code-owner review enforcement and admin enforcement remain off so the repository does not become unmergeable.
- Tighten code-owner review enforcement, admin bypass, and maintainer isolation once a second maintainer can satisfy the review requirement.

Untrusted pull request policy:

- Pull request workflows must stay fork-safe and use read-only `contents` permission unless a specific trusted workflow boundary justifies more.
- Do not expose release, package, deployment, or private infrastructure secrets to code from untrusted forks.
- Avoid `pull_request_target` for workflows that check out or execute contributor-controlled code.
- Dependency review, policy checks, formatting, linting, generated-code drift checks, and tests are acceptable untrusted PR validation surfaces when they run without secrets.

Maintainer and automation isolation:

- Prefer `GITHUB_TOKEN` with explicit job-scoped permissions over personal access tokens or long-lived bot credentials.
- Use a dedicated low-privilege bot identity only when new automation needs privileges that `GITHUB_TOKEN` cannot safely provide.
- Keep release and deployment writes limited to trusted branch or tag workflows with explicit job-scoped permissions.
- Enable code-owner review enforcement, stricter admin bypass policy, and stronger review separation only after at least two maintainers can satisfy those controls without blocking routine maintenance.
- Do not use self-hosted runners for untrusted pull requests. Only consider self-hosted runners for trusted branches or tags after documenting isolation, secret exposure, cleanup, and network-access controls.

When changing GitHub settings, keep the repo aligned with:

- [SECURITY.md](../SECURITY.md) for disclosure handling and patch timing.
- [docs/security-operations.md](security-operations.md) for recurring security operations and deeper testing expectations.
- [.github/CODEOWNERS](../.github/CODEOWNERS) for sensitive file ownership.
- [.github/workflows/pull-request.yml](../.github/workflows/pull-request.yml) for fork-safe checks.
- [.github/workflows/release.yml](../.github/workflows/release.yml) for release publication, checksums, SBOMs, and attestations.

## Release publication

Public releases use annotated semantic version tags in the `v0.y.z` range. The
release contract is a tagged release for the `openhealth` binary, the
single-file OpenHealth skill, and the direct local runtime. Tag a version like
`v0.3.0`, push the tag, and let the release workflow:

- validate storage codegen, formatting, and tests before publish
- build binaries with `openhealth --version` set from the tag
- require `docs/release-notes/<tag>.md` and a matching `CHANGELOG.md` entry
  before publishing
- create or reuse only a draft GitHub Release before assets are attached
- use `docs/release-notes/<tag>.md`, for example
  `docs/release-notes/v0.3.0.md`, as the GitHub Release body
- keep release-note paragraphs and list items on one source line so GitHub
  Releases and API clients do not show hard-wrapped prose
- attach platform binary archives, the skill archive, the canonical source
  archive, release installer, SHA256 checksums, and SPDX SBOM
- verify the draft release has the expected asset set before publication
- generate GitHub attestations for the published assets
- publish the draft only after all assets and attestations are ready, then
  verify the release is latest

The tag push is the maintainer approval boundary for release publication. After `verify-release` passes, the publish job runs without a separate GitHub environment approval prompt.

Before tagging, add `docs/release-notes/<tag>.md`, update `CHANGELOG.md`, and run `./scripts/validate-release-docs.sh <tag>` locally. The release workflow runs the same check before publishing and does not fall back to generated GitHub release notes.

After this draft-first workflow is active, enable GitHub release immutability in repository settings for future releases. Published release tags and assets should then be treated as immutable; fix bad artifacts with a new patch release instead of replacing assets on an existing release. Existing releases such as `v0.3.0` are left as already published.
