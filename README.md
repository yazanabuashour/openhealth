# openhealth

Open source project scaffold with maintainer task tracking managed through Beads.

## For Contributors

Beads is optional for outside contributors and is not required to open, review, or merge PRs.

## For Maintainers

This repo uses **Beads** (`bd`) in embedded mode for maintainer task tracking.

Preferred tool install:

```bash
mise install
```

Alternative:

```bash
brew install beads dolt
```

Second machine bootstrap:

```bash
git clone git@github.com:yazanabuashour/openhealth.git
cd openhealth
bd bootstrap
bd hooks install
git config beads.role maintainer
bd list
```

Sync between machines:

```bash
bd dolt push
bd dolt pull
```

If `bd dolt pull` reports uncommitted Dolt changes:

```bash
bd dolt commit
bd dolt pull
```
