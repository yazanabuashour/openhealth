# openhealth

## Maintainer Task Tracking

This repository uses **Beads** (`bd`) for maintainer task tracking in **embedded mode**.
Beads is optional for outside contributors and is not required to open or review PRs.

### Maintainer Setup

Install these tools on each maintainer machine:

```bash
brew install beads dolt
bd version
dolt version
```

This repository is initialized with:

```bash
bd init --prefix oh
bd dolt remote add origin git+ssh://git@github.com/yazanabuashour/openhealth.git
```

### Second Machine Bootstrap

Clone the repo normally, then bootstrap the Beads database from the GitHub remote:

```bash
git clone git@github.com:yazanabuashour/openhealth.git
cd openhealth
bd bootstrap
bd hooks install
bd list
```

If Beads warns that `beads.role` is not configured in a maintainer clone, set it once:

```bash
git config beads.role maintainer
```

### Daily Sync

Before leaving one machine:

```bash
bd dolt push
```

After arriving on another machine:

```bash
bd dolt pull
```

If `bd dolt pull` reports uncommitted Dolt changes:

```bash
bd dolt commit
bd dolt pull
```

### Contributor Policy

External contributors may ignore Beads entirely. CI, review, and merge decisions must not depend on reviewer access to local Beads state.
