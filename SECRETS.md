Secrets & config handling (Momo CLI)

Principles:
- Never commit API keys, private keys, or passphrases to source control.
- Store secrets in CI secrets (GitHub Actions secrets) or a dedicated secrets manager (Vault, AWS Secrets Manager).
- Use environment variables at runtime; prefer OS-level service configuration for production.

Recommended variables (CI & runtime):
- PROVIDER_API_KEY: API key for the LLM provider (store as secret)
- GPG_PRIVATE_KEY: base64 ASCII-armored private key for CI signing
- GPG_PASSPHRASE: passphrase to unlock private key
- DOCKER_REGISTRY_TOKEN: docker push/pull token (if using container registry)

Local development:
- Use .env file (gitignored) or set env in shell. Provide .env.example with placeholders only.

Example (PowerShell):
  $env:PROVIDER_API_KEY = "sk_..."
  $env:GPG_PASSPHRASE = "my-passphrase"
  .\momo.exe

CI (GitHub Actions):
- Add secrets in repository Settings → Secrets → Actions. Reference as ${{ secrets.PROVIDER_API_KEY }}.
