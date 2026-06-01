# Roadmap

`codex-action-guard` is intentionally narrow in v0: safe-by-default workflows for `openai/codex-action` and direct `codex exec` in GitHub Actions.

## v0.1

- Improve source position precision for shell-generated prompt files.
- Add more tests for workflow-level versus job-level permissions.
- Tune `CODX005` and `CODX010` against real-world safe posting patterns.
- Publish installable release artifacts.

## v0.2

- Add a machine-readable rule metadata export.
- Add more profile customization without weakening safe defaults.
- Improve review packets for changed workflow diffs.
- Add examples for GitHub code scanning SARIF upload.

## Later

- Consider additional Codex GitHub Action provider-pack rules after v0 false-positive data improves.
- Consider a stable JSON schema document for report consumers.

## Non-goals for v0

- Claude, Gemini, MCP, GitLab, Jenkins, Azure Pipelines, n8n, browser automation, or generic AI workflow scanning.
- SaaS features.
- LLM-backed analysis by default.
- A generic AI code reviewer.
