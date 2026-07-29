Production deploy (Momo CLI)

Goal:
- Release signed installers/binaries, publish to GitHub Releases, and deploy images to container registry.

Steps:
1. Prepare release
   - Ensure CI secrets are set: GPG_PRIVATE_KEY (base64), GPG_PASSPHRASE, GITHUB_TOKEN, optional registry creds.
   - Tag commit: git tag vX.Y.Z && git push origin vX.Y.Z
2. CI will run goreleaser and upload artifacts to GitHub Releases (archives, deb/rpm, SBOM, signatures).
3. Pull artifacts to target machines or use package manager (dpkg/rpm) where supported.
4. For containerized usage, pull image from your registry and run: docker run --rm your-registry/momo:latest

Rollback:
- Use previous release tag. If compromise suspected, revoke old GPG key and rotate.

Security:
- Use private repos and protect release workflow with required reviewers if needed.
- Do not store secrets in repo. Use GitHub Secrets or a secrets manager.
