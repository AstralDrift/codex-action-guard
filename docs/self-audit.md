# Self-audit

This repository scans its own GitHub Actions workflows in `.github/workflows/self-audit.yml`.

Current sample from this repository:

```markdown
# codex-action-guard audit report

- Tool: `codex-action-guard dev`
- Rule version: `v0`
- Root: `/path/to/codex-action-guard`
- Scanned files: `3`
- Codex workflow files: `0`

## Summary counts

- critical: 0
- high: 0
- medium: 0
- low: 0
- info: 0

## Not applicable / no Codex workflows found

No `openai/codex-action` or direct `codex exec` invocations were found in scanned workflow files.

## Profile suggestions

- No Codex workflows found. Start with `codex-action-guard init --profile pr-review-readonly` for a safe read-only profile.
```

If this project adds Codex-powered workflows later, this document should be updated with current output and any accepted review-required findings.
