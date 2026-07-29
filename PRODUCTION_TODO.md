MOMO — Production readiness checklist (Hinglish)
===============================================

Status — Kya ho chuka (Completed)
- Project scaffold, cmd, internal/* implemented and builds (go build).
- Provider: OpenAI-compatible HTTP client + streaming (SSE) implemented.
- Agent loop: message history, streaming hook, tool-call loop implemented.
- Tools: read_file, list_dir, run_cmd, write_file, edit_file, search, web_fetch, apply_patch (apply mode implemented).
- Executor: permission flow + TUI permission dialog with queueing and approve/remember.
- Approvals persisted (global + project: ~/.config/momo/config.json and workspace/.momo/config.json).
- Security docs: SECURITY.md, RELEASE.md.
- CI workflow + Dependabot; added CI job to run internal/tools tests on Linux runner.
- Unit tests for agent/tools/config added (some local tests blocked by Windows App Control; CI runner will run all tests on Linux).
- Closed-source setup: module renamed to momo, PRIVATE.md and All-rights LICENSE added.
- Publish helper script (scripts/publish.ps1) created.

Priority remaining work (to make 100% production-ready)
-------------------------------------------------------
High priority (must before public/private release):
1) CI signing & SBOM — build artifacts, create SBOM, sign binaries (GPG) in CI; upload signed artifacts to releases. (2-3 days)
2) Secure self-update implementation — CLI can check private release metadata, download signed artifact, verify signature, replace binary with rollback support. (2-3 days)
3) apply_patch audit & tests — security review and more tests for complex patches & path sanitization. (2-4 days)
4) Provider adapters hardening — per-provider error handling, retries, rate-limits, and auth variations. (2-3 days)

Medium priority:
5) Packaging — produce signed installers (MSI), deb/rpm, multi-arch binaries, and Docker images. (2 days)
6) Observability & runbook — logging, metrics endpoint, health checks, Sentry/crash reporting with redaction. (2 days)
7) CI secrets & protected release process — ensure GPG keys and provider keys stored securely. (1 day)

Low priority / nice-to-have:
8) TUI polish & UX — better dialogs, permission history UI, per-project allowlist editor. (2 days)
9) More unit/integration tests — provider mocks, streaming tests, long-run smoke tests. (ongoing)
10) Documentation — RELEASE.md, INSTALL.md, SECURITY.md expanded. (1 day)

How to run dev & tests locally
------------------------------
- Build: cd D:\boxcode\momo && go build -o momo.exe ./cmd
- Run: .\momo.exe  (interactive TUI)
- Unit tests: go test ./...  (If Windows App Control blocks test exe, run tests in WSL or CI on Ubuntu: go test ./internal/tools -v)

Next immediate actions completed:
- Implemented safe self-update CLI (downloads, GPG verify with provided public key, atomic replace on Unix, .new instruction on Windows).
- Created goreleaser config and GitHub release workflow to produce multi-arch artifacts, SBOM, and optional signing when CI secrets provided.
- Hardened apply_patch with transactional apply, backups, symlink/path checks, and expanded tests (all pass locally).
- Added reproducible build flags (Makefile), nfpm deb/rpm packaging in goreleaser config, Dockerfile for container image, and a monitoring stub + runbook.

When 100% done I will: create a final DONE file listing all completed items and steps to release.


Notes for you (Hinglish):
- Repo can stay closed-source on GitHub by creating a private repo (use scripts/publish.ps1).
- For update notifications, CLI should check private release metadata with a token and show TUI banner.
- Signing + verification are crucial for safe self-update — don't skip.

