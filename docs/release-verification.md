# Release Verification

Tagged OpenHealth releases publish integrity-focused assets alongside the Git
tag:

- `openhealth_<version>_<os>_<arch>.tar.gz`
- `openhealth_<version>_skill.tar.gz`
- `openhealth_<version>_source.tar.gz`
- `openhealth_<version>_checksums.txt`
- `openhealth_<version>_sbom.spdx.json`
- `install.sh`

The platform archives contain the production `openhealth` binary. The
skill archive contains the single shipped `SKILL.md` payload. The source archive
is the canonical source artifact for the Go module and local runtime. The
installer script downloads and verifies the matching platform archive before
installing the same-tag runner. It then prints the required second step:
register the same-tag skill source or archive with the user's agent using that
agent's native skill system. The skill archive is the portable release artifact
for agents that install from files instead of GitHub paths. The checksums file and
GitHub attestations let users verify that the assets were produced by this
repository's release workflow.

## Verify a release

Download the assets from the GitHub Release page for the tag you want to verify,
then run:

```bash
shasum -a 256 -c openhealth_<version>_checksums.txt
gh attestation verify openhealth_<version>_<os>_<arch>.tar.gz --repo yazanabuashour/openhealth
gh attestation verify openhealth_<version>_skill.tar.gz --repo yazanabuashour/openhealth
gh attestation verify openhealth_<version>_source.tar.gz --repo yazanabuashour/openhealth
gh attestation verify install.sh --repo yazanabuashour/openhealth
```

If these commands succeed, the assets match the published checksums and have
valid GitHub attestations for this repository.

## SBOM

The SPDX JSON SBOM asset is intended for audit tooling and manual inspection:

```bash
jq '.packages | length' openhealth_<version>_sbom.spdx.json
```

The SBOM is generated from the tagged source contents during the release
workflow and attached to the same GitHub Release as the binary, skill, and
source archives.
