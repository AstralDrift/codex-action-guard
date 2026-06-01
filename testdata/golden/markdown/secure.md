# codex-action-guard audit report

- Tool: `codex-action-guard 1.0.0`
- Rule version: `v0`
- Root: `<ROOT>`
- Scanned files: `3`
- Codex workflow files: `1`

## Summary counts

- critical: 0
- high: 0
- medium: 0
- low: 0
- info: 0

## Findings

No findings.

## Safe patterns found

- `.github/workflows/codex-secure.yml:24`: Codex uses a checked-in prompt-file instead of large inline prompt text.
- `.github/workflows/codex-secure.yml:24`: Codex output is constrained by an output schema.
- `.github/workflows/codex-secure.yml:24`: Codex runs with read-only sandbox settings.
- `.github/workflows/codex-secure.yml:16`: Codex job declares read-only or empty GITHUB_TOKEN permissions.
- `.github/workflows/codex-secure.yml:23`: actions/checkout disables persisted credentials.
- `.github/workflows/codex-secure.yml:53`: Codex uses a checked-in prompt-file instead of large inline prompt text.
- `.github/workflows/codex-secure.yml:53`: Codex output is constrained by an output schema.
- `.github/workflows/codex-secure.yml:53`: Codex runs with read-only sandbox settings.
- `.github/workflows/codex-secure.yml:52`: actions/checkout disables persisted credentials.
- `.github/workflows/codex-secure.yml:42`: Codex job has an obvious trusted gate such as actor, label, allow-users, or environment approval.

