Security & Vulnerability Reporting
=================================

Report vulnerabilities
---------------------
If you find a security issue, please email: security@example.com (replace with real contact) with details.

Quick security checklist (pre-release)
-------------------------------------
- Do NOT commit API keys or secrets.
- Use GitHub Secrets / Vault for CI and runtime secrets.
- Run dependency scans (Dependabot + `govulncheck`).
- Run static analysis (gofmt, go vet, go vet -vettool, semgrep).
- Enforce least-privilege for tools: read-only by default; require explicit allow for write/exec.
- Truncate tool outputs and never log secrets.
- Redact secrets in crash reports before sending.

Release gating
--------------
- All findings must be fixed or accepted and documented before publishing a signed release.

Acknowledgements
----------------
Thanks for helping keep this project secure.