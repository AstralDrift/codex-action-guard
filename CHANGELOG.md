# Changelog

This project follows Semantic Versioning once tagged releases begin. Changes are grouped using Conventional Commit categories.

## Unreleased

### Added

- Initial Go CLI with `version`, `init`, `audit`, `diff`, `packet`, and `explain`.
- v0 OpenAI Codex GitHub Action provider rule pack, `CODX001` through `CODX010`.
- Safe generated profiles for read-only PR review, CI failure analysis, release notes drafting, security review, and label-gated maintainer tasks.
- Markdown, JSON, and SARIF report output.
- Deterministic `rules` metadata export for downstream tooling.
- JSON schemas for audit reports and rule metadata exports.
- Tag-triggered release workflow that publishes cross-platform CLI archives and checksums.
- CI dogfooding of the repository's own workflows.
