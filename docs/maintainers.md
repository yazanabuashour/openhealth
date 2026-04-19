# Maintainer Notes

This repository uses **Beads** (`bd`) in embedded mode for maintainer task tracking.

This repository is public and includes a production `openhealth` runner binary,
a single-file OpenHealth skill, an optional direct-local Go SDK, and a local
SQLite runtime. Keep maintainer docs honest about the actual supported surface.

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
- Release publication runs in a protected `release` environment with narrowly scoped write permissions.
- Security reports are expected through GitHub private vulnerability reporting.

Current review enforcement nuance:

- The repository currently has a single maintainer account.
- `main` requires pull requests, status checks, conversation resolution, and one approving review, but code-owner review enforcement and admin enforcement remain off so the repository does not become unmergeable.
- Tighten code-owner review enforcement, admin bypass, and maintainer isolation once a second maintainer can satisfy the review requirement.

When changing GitHub settings, keep the repo aligned with:

- [SECURITY.md](../SECURITY.md) for disclosure handling and patch timing.
- [.github/CODEOWNERS](../.github/CODEOWNERS) for sensitive file ownership.
- [.github/workflows/pull-request.yml](../.github/workflows/pull-request.yml) for fork-safe checks.
- [.github/workflows/release.yml](../.github/workflows/release.yml) for release publication, checksums, SBOMs, and attestations.

## Release publication

Public releases use annotated semantic version tags in the `v0.y.z` range. The
release contract is a tagged release for the `openhealth` binary, the
single-file OpenHealth skill, and the direct local runtime. Tag a version like
`v0.2.1`, push the tag, and let the release workflow:

- validate storage codegen, formatting, and tests before publish
- build binaries with `openhealth --version` set from the tag
- create or reuse the GitHub Release
- use `docs/release-notes/<tag>.md` when present, for example
  `docs/release-notes/v0.2.2.md`, as the GitHub Release body
- keep release-note paragraphs and list items on one source line so GitHub
  Releases and API clients do not show hard-wrapped prose
- attach platform binary archives, the skill archive, the canonical source
  archive, release installer, SHA256 checksums, and SPDX SBOM
- generate GitHub attestations for the published assets

The `release` environment should remain protected so only approved maintainers can publish release assets.
