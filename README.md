# openhealth

`openhealth` is a public project scaffold. It currently exists to establish the repository policy, maintainer workflow, and release hygiene for an eventual open source project.

## Current status

This repository does not yet ship a runnable app, package, or published API. The current `main` branch is a maintained scaffold for contributor guidance, repository automation, and maintainer operations.

- Stable today: repository policies, contributor workflow, maintainer workflow notes, and GitHub release scaffolding.
- Not stable yet: runtime behavior, public APIs, package contracts, deployment surfaces, and backward-compatibility guarantees.

Until the project reaches `0.1.0`, compatibility is best effort and may change between releases.

## Repository contents

- [CONTRIBUTING.md](CONTRIBUTING.md) explains how outside contributors should propose changes.
- [SECURITY.md](SECURITY.md) explains how to report vulnerabilities privately and what response timing to expect.
- [docs/maintainers.md](docs/maintainers.md) documents Beads-based maintainer workflow and repo administration notes.
- [LICENSE](LICENSE) defines the project license.

## Release contract

The initial release surface is GitHub Releases with semantic version tags in the `0.y.z` range. Release notes are generated from protected tags. This repository does not currently publish packages or downloadable build artifacts.

## Data and credentials

This repository is intended to be safe to view publicly.

- No production credentials or private infrastructure secrets should be committed here.
- No production datasets or personal data are included here.
- Outside contributors are not expected to use Beads, Dolt, or private maintainer state to open or review pull requests.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution expectations and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community standards.
