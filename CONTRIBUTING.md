# Contributing

## Beads Policy

This repository uses **Beads** (`bd`) as a maintainer workflow tool.

- Outside contributors are not required to install or use Beads.
- Pull requests must be reviewable without access to Beads state.
- CI, checks, and merge gates must not depend on Beads.

If you are contributing code, focus on the code changes, tests, and documentation required for your PR. Maintainers may separately use Beads to track planning and follow-up work.

## Maintainer Notes

Maintainers use Beads in embedded mode with the repository GitHub remote as the Dolt remote.

```bash
bd init --prefix oh
bd dolt remote add origin git+ssh://git@github.com/yazanabuashour/openhealth.git
```

On a second machine:

```bash
git clone git@github.com:yazanabuashour/openhealth.git
cd openhealth
bd bootstrap
bd hooks install
```

If a maintainer clone warns about role detection, set:

```bash
git config beads.role maintainer
```

When switching machines:

```bash
bd dolt push
bd dolt pull
```
