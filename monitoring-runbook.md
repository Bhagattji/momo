Monitoring & Runbook (Momo CLI)

Goal:
- Basic observability for CLI when run on servers or in background: structured logs, optional metrics push, and crash reporting.

What to collect:
- Structured logs (INFO/WARN/ERROR) with operation context and request IDs.
- Metrics: command invocation counts, tool-execution durations, error rates.
- Crash reports: upload stack traces (redact secrets) to a private Sentry/Crash service.

Implementation steps:
1. Integrate logging library (logrus/zerolog) and ensure all components use it.
2. Add OpenTelemetry SDK for metrics and traces (optional): export to OTLP collector.
3. Add a small background telemetry sender for CLI runs longer than N seconds.
4. Add crash-reporting hook with manual opt-in for privacy.

Runbook (when alert fires):
- Check recent releases and deployment status.
- Pull logs from affected hosts or CI artifacts.
- If crash related to update/self-update: disable auto-update rollouts, revoke GPG keys if compromised.

Notes:
- Do not send secrets or full file contents in telemetry.
- Provide an opt-out flag for telemetry.
