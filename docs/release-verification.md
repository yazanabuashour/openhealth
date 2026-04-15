# Release Verification

OpenHealth releases publish three integrity-focused assets alongside the Git tag:

- `openhealth_<version>_source.tar.gz`
- `openhealth_<version>_checksums.txt`
- `openhealth_<version>_sbom.spdx.json`

The source archive is the canonical release artifact for the local in-process runtime. The checksums file and GitHub attestations let users verify that the archive was produced by this repository's release workflow.

## Verify a release

Download the three assets from the GitHub Release page for the tag you want to verify, then run:

```bash
shasum -a 256 -c openhealth_<version>_checksums.txt
gh attestation verify openhealth_<version>_source.tar.gz --repo yazanabuashour/openhealth
```

If both commands succeed, the archive matches the published checksum and the artifact has a valid GitHub attestation for this repository.

## SBOM

The SPDX JSON SBOM asset is intended for audit tooling and manual inspection:

```bash
jq '.packages | length' openhealth_<version>_sbom.spdx.json
```

The SBOM is generated from the tagged source contents during the release workflow and attached to the same GitHub Release as the source archive.
