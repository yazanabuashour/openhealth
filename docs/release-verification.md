# Release Verification

Tagged OpenHealth releases publish:

- `openhealth_<version>_<os>_<arch>.tar.gz`
- `openhealth_<version>_skill.tar.gz`
- `openhealth_<version>_source.tar.gz`
- `openhealth_<version>_checksums.txt`
- `openhealth_<version>_sbom.spdx.json`
- `install.sh`

The platform archives contain the production `openhealth` binary. The skill
archive contains the shipped `SKILL.md`. The source archive is the canonical Go
module and local runtime source artifact.

The installer verifies the matching platform archive, installs the same-tag
runner, prints `openhealth --version`, and tells users to register the same-tag
skill source or archive with their agent. Checksums and GitHub attestations
verify that release assets were produced by this repository's workflow.

## Verify a Release

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

## Smoke-Test an Install

Install into a temporary directory, then verify the runner version and domains:

```bash
install_dir="$(mktemp -d)"
OPENHEALTH_INSTALL_DIR="$install_dir" \
  OPENHEALTH_VERSION=v0.2.2 \
  sh -c "$(curl -fsSL https://github.com/yazanabuashour/openhealth/releases/download/v0.2.2/install.sh)"

export PATH="$install_dir:$PATH"
command -v openhealth
openhealth --version
openhealth --help
```

The valid runner domains are `weight`, `blood-pressure`, `medications`, `labs`,
`body-composition`, and `imaging`.

## SBOM

The SPDX JSON SBOM asset is intended for audit tooling and manual inspection:

```bash
jq '.packages | length' openhealth_<version>_sbom.spdx.json
```

The SBOM is generated from the tagged source contents during the release
workflow and attached to the same GitHub Release as the binary, skill, and
source archives.
