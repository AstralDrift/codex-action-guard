# Roadmap

`codex-action-guard` is intentionally narrow in v0: safe-by-default workflows for `openai/codex-action` and direct `codex exec` in GitHub Actions.

## v0.1

- Ship the first `v0.1.0` release and maintain the floating `v0` action tag.
- Add a first-class `install` command for safe guard workflow adoption.
- Keep polishing docs and examples for artifact and SARIF onboarding.
- Improve source position precision for shell-generated prompt files.
- Add more tests for workflow-level versus job-level permissions.
- Tune `CODX005` and `CODX010` against real-world safe posting patterns.

## v0.2

- Add more profile customization without weakening safe defaults.
- Improve review packets for changed workflow diffs.
- Add more examples for gated write workflows.

## Later

- Consider additional Codex GitHub Action provider-pack rules after v0 false-positive data improves.
- Consider broader downstream integrations once v0 report and rule schemas settle.

## Completed toward v0.1

- Deterministic machine-readable rule metadata export.
- JSON schemas for audit reports and rule metadata exports.
- Tag-triggered release workflow that publishes cross-platform CLI archives and checksums.
- Examples for artifact and SARIF guard workflow adoption.

## Non-goals for v0

- Claude, Gemini, MCP, GitLab, Jenkins, Azure Pipelines, n8n, browser automation, or generic AI workflow scanning.
- SaaS features.
- LLM-backed analysis by default.
- A generic AI code reviewer.
