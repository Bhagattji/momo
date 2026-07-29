Momo CLI — Release Runbook

Scope
- Build signed multi-arch binaries, create SBOM (CycloneDX), sign artifacts, upload to GitHub Releases, and enable self-update verification.

Prerequisites
- Private GitHub repository with Release permissions.
- CI secrets:
  - GITHUB_TOKEN (Actions default or PAT with repo:packages, releases)
  - GPG_PRIVATE_KEY (base64-encoded ASCII-armored private key)
  - GPG_PASSPHRASE (passphrase for private key)
- (Optional) Provider API keys for integration tests

CI configuration (what was added)
- .github/workflows/release.yml — triggers on tag push (v*). Generates SBOM via syft, runs goreleaser.
- .goreleaser.yml — builds linux/darwin/windows for amd64/arm64, produces archives and signs with gpg when secrets present.
- .github/workflows/ci.yml — build + test, and tools-test job on ubuntu to avoid Windows App Control.

How to add secrets
1. Open repository Settings → Secrets → Actions.
2. Add GPG_PRIVATE_KEY: base64 of `gpg --export-secret-keys --armor <KEYID> | base64`.
3. Add GPG_PASSPHRASE: the private key passphrase.
4. Ensure GITHUB_TOKEN has permission to create releases (Actions default works for same repo).

How release works
1. Create a git tag: git tag vX.Y.Z && git push origin vX.Y.Z
2. Release workflow runs: generates SBOM, calls goreleaser to build artifacts, sign (if secrets exist), and upload to GitHub Releases.
3. Artifacts: binaries, archives, sbom-cyclonedx.json, and detached signatures (.sig) when signing enabled.

Verifying artifacts locally
- Download binary and .sig
- Import public key: gpg --import pubkey.asc
- Verify: gpg --verify momo.sig momo
- Inspect SBOM: jq . sbom-cyclonedx.json or use CycloneDX viewers

Self-update
- CLI `momo self-update` can fetch release metadata (GitHub API) and download chosen asset and signature.
- For verification, set env GPG_PUBLIC_KEY to ASCII-armored public key or pass --pubkey file to the command.
- On success, CLI replaces binary atomically on Unix; on Windows it writes .new and instructs manual replace.

Rollbacks & safety
- goreleaser releases are immutable per tag; keep keys offline and rotate if compromised.
- Self-update verifies signatures before replacing; backups are kept by apply_patch and self-update keeps a .bak when replacing.

Next manual steps for you (to finish):
- Add CI secrets as above.
- Replace ORG in .goreleaser.yml with your GitHub org/user and repo name if needed, or set env GITHUB_REPO in CI.
- (Optional) Add MSI packaging steps to goreleaser config if you need installers.

Contact
- When secrets are added and a tag is pushed, CI will produce signed releases. I will verify and finalize DONE notes.
