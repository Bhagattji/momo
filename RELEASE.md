Release & Signing
=================

Steps to make a production release:
1. Ensure CI (tests, vet, lint) pass on main branch.
2. Generate SBOM (CycloneDX/SPDX) for the build.
3. Build reproducible static binaries for supported platforms.
4. Sign binaries and attach checksums + SBOM to the GitHub Release.
5. Verify signatures and upload artifacts to release and container registry.

Verification commands (developer machine):
  go test ./...
  go vet ./...
  # build: go build -trimpath -o matha.exe ./cmd

Document any deviations in RELEASE_NOTES.md before publishing.