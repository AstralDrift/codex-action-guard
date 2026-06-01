# codex-action-guard

Safe-by-default Codex GitHub Action workflows.

[![CI](https://github.com/AstralDrift/codex-action-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/AstralDrift/codex-action-guard/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/AstralDrift/codex-action-guard?include_prereleases)](https://github.com/AstralDrift/codex-action-guard/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/AstralDrift/codex-action-guard.svg)](https://pkg.go.dev/github.com/AstralDrift/codex-action-guard)

`codex-action-guard` is an independent community project. It is not affiliated with, endorsed by, or certified by OpenAI.

Codex in GitHub Actions is powerful, but unsafe workflow composition can put prompts, secrets, write tokens, untrusted PR/comment/issue text, and downstream shell or API actions in the same trust boundary. This tool helps maintainers generate safe Codex workflow profiles and audit existing workflows that use `openai/codex-action` or direct `codex exec`.

## Project status

This project is in early v0 development. The CLI is usable, deterministic, and tested, but rule tuning will keep improving as maintainers share real workflow shapes and false-positive reports.

## What it does

- Generates safe Codex GitHub Action workflow profiles.
- Audits `.github/workflows/*.yml` and `.github/workflows/*.yaml`.
- Inspects Codex prompt and schema files when relevant.
- Emits evidence-bound findings in Markdown, JSON, and SARIF 2.1.0.
- Builds structured review packets for humans or Codex.
- Explains each rule with examples, remediation, and false-positive notes.

This is intentionally not a broad agentic workflow scanner. v0 focuses on the OpenAI Codex GitHub Action provider pack.

## Quick start

Install the guard in your repo:

```sh
codex-action-guard install --preset artifact
```

Generate a safe Codex profile:

```sh
codex-action-guard init --profile pr-review-readonly
```

Audit manually:

```sh
codex-action-guard audit --all --fail-on high
```

## Install

```sh
go install github.com/AstralDrift/codex-action-guard/cmd/codex-action-guard@latest
```

For local development:

```sh
go test ./...
go run ./cmd/codex-action-guard audit --all
```

## Commands

```sh
codex-action-guard version
codex-action-guard install --preset artifact
codex-action-guard install --preset sarif --out ../target-repo
codex-action-guard init --profile pr-review-readonly
codex-action-guard audit --all --format markdown
codex-action-guard audit .github/workflows/codex.yml --format sarif --output codex-action-guard.sarif
codex-action-guard diff main...HEAD --fail-on high
codex-action-guard packet --target human --changed main...HEAD
codex-action-guard rules --format json
codex-action-guard explain CODX001
```

## GitHub Action usage

`codex-action-guard install` writes one of the workflows below for you. The examples use `AstralDrift/codex-action-guard@v0`, a floating major tag maintained by the release process and backed by the latest compatible v0 release.

```yaml
name: Codex Action Guard

on:
  pull_request:
    paths:
      - ".github/workflows/**"
      - ".github/codex/**"
      - "AGENTS.md"
      - "action.yml"
  workflow_dispatch:

permissions:
  contents: read
  security-events: write

jobs:
  codex-action-guard:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
        with:
          persist-credentials: false

      - uses: AstralDrift/codex-action-guard@v0
        with:
          fail-on: high
          format: sarif
          output: codex-action-guard.sarif

      - uses: github/codeql-action/upload-sarif@v4
        if: always()
        with:
          sarif_file: codex-action-guard.sarif
```

The action wrapper currently uses Go from the checked-out action source. Tagged releases also publish CLI archives and checksums for direct installation.

See [docs/install.md](docs/install.md) for the artifact and SARIF installer presets.

## Safe profiles

The `init` command writes a workflow, prompt, output schema, and threat-model document. It refuses to overwrite existing files unless `--force` is set, then validates generated output by running the internal audit.

Available profiles:

- `pr-review-readonly`
- `ci-failure-analysis-readonly`
- `release-notes-draft`
- `security-review-readonly`
- `label-gated-maintainer-task`

See [docs/profiles.md](docs/profiles.md) for the threat model and generated files for each profile.

## Rule pack

v0 ships ten Codex-specific rules:

- `CODX001`: Untrusted GitHub event content reaches Codex prompt.
- `CODX002`: Untrusted content reaches Codex prompt in write-capable job.
- `CODX003`: `OPENAI_API_KEY` or `CODEX_API_KEY` exposed at job scope with repo-controlled code.
- `CODX004`: Codex uses `danger-full-access` or unsafe strategy without trusted trigger or gate.
- `CODX005`: Codex output feeds shell/github-script/gh/deploy without schema validation.
- `CODX006`: `pull_request_target` or `workflow_run` checks out untrusted code before Codex.
- `CODX007`: Codex job has broad `GITHUB_TOKEN` permissions.
- `CODX008`: prompt-file or schema file modified in the same PR that triggers Codex.
- `CODX009`: write-capable Codex workflow lacks actor, allow-users, maintainer label, or environment gate.
- `CODX010`: Codex output is posted without size limits, escaping, redaction, or schema constraints.

Findings use "unsafe trust boundary" and "review required" language unless the rule can show a concrete source-to-boundary-to-sink path.

See [docs/rules.md](docs/rules.md) and `codex-action-guard explain <RULE_ID>` for rule details.

## Example finding

```text
CODX001: Untrusted GitHub event content reaches Codex prompt
Location: .github/workflows/codex.yml:22
Source: comment body
Prompt boundary: with.prompt
Privilege context: write permissions: issues
Safer pattern: Use a trusted prompt-file, pass stable identifiers, sanitize untrusted text, or require a maintainer gate.
```

## Output formats

Markdown reports are designed for maintainers. JSON reports provide a stable schema with metadata, scanned files, findings, detected invocations, safe patterns, and profile suggestions. SARIF output is suitable for code scanning ingestion.

See [docs/report-formats.md](docs/report-formats.md) for schema notes and examples.

## Machine-readable outputs

Downstream tooling can consume:

- JSON audit reports, described by [`schemas/report.schema.json`](schemas/report.schema.json).
- Rule metadata from `codex-action-guard rules --format json`, described by [`schemas/rules.schema.json`](schemas/rules.schema.json).
- SARIF 2.1.0 reports for GitHub code scanning upload.

## Documentation

- [Architecture](docs/architecture.md)
- [Usage guide](docs/usage.md)
- [Install guide](docs/install.md)
- [Profiles](docs/profiles.md)
- [Rule reference](docs/rules.md)
- [Report formats](docs/report-formats.md)
- [Threat model](docs/threat-model.md)
- [Safe patterns](docs/safe-patterns.md)
- [Comparison](docs/comparison.md)
- [Rule design](docs/rule-design.md)
- [Self-audit](docs/self-audit.md)
- [Release process](docs/release.md)
- [Roadmap](ROADMAP.md)

## Comparison

`codex-action-guard` complements actionlint, zizmor, CodeQL, broad workflow scanners, AI code reviewers, and prompt eval tools. It focuses narrowly on safe Codex GitHub Action profiles and Codex-specific workflow trust-boundary analysis. See [docs/comparison.md](docs/comparison.md).

## Non-goals

v0 does not support Claude, Gemini, MCP, GitLab, Jenkins, Azure Pipelines, n8n, browser automation, generic AI workflows, SaaS features, or LLM-backed analysis by default.

## Development

```sh
go test ./...
go run ./cmd/codex-action-guard audit --all --fail-on high
```

The repository dogfoods the scanner in CI by auditing its own GitHub Actions workflows.

## Contributing and security

Contributions are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md), especially the rule quality bar and false-positive expectations.

Please do not report suspected vulnerabilities in public issues. See [SECURITY.md](SECURITY.md).
