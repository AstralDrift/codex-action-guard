# Changelog

This project follows Semantic Versioning once tagged releases begin. Changes are grouped using Conventional Commit categories.

## Unreleased

No unreleased changes yet.

## v0.1.1 - 2026-06-01

### Changed

- Generated workflows now use `actions/upload-artifact@v5`.
- Generated workflows now opt into the GitHub Actions Node 24 JavaScript action runtime.
- Workflows and generated templates now use `actions/checkout@v6`.
- Release documentation now uses a reusable tag-and-floating-major flow.

## v0.1.0 - 2026-06-01

### Added

- Initial Go CLI with `version`, `init`, `audit`, `diff`, `packet`, and `explain`.
- v0 OpenAI Codex GitHub Action provider rule pack, `CODX001` through `CODX010`.
- Safe generated profiles for read-only PR review, CI failure analysis, release notes drafting, security review, and label-gated maintainer tasks.
- Markdown, JSON, and SARIF report output.
- Deterministic `rules` metadata export for downstream tooling.
- JSON schemas for audit reports and rule metadata exports.
- Tag-triggered release workflow that publishes cross-platform CLI archives and checksums.
- `install` command with artifact and SARIF guard workflow presets.
- Generated workflow examples and installer documentation.
- CI dogfooding of the repository's own workflows.

### Changed

- Release workflow now runs only for semver-like `vMAJOR.MINOR.PATCH` tags so the floating `v0` action tag does not create a separate release.
