# codex-action-guard

Safe-by-default Codex GitHub Action workflows.

`codex-action-guard` is an independent community project. It is not affiliated with, endorsed by, or certified by OpenAI.

Codex in GitHub Actions is powerful, but unsafe workflow composition can put prompts, secrets, write tokens, untrusted PR/comment/issue text, and downstream shell or API actions in the same trust boundary. This tool helps maintainers generate safe Codex workflow profiles and audit existing workflows that use `openai/codex-action` or direct `codex exec`.

## What it does

- Generates safe Codex GitHub Action workflow profiles.
- Audits `.github/workflows/*.yml` and `.github/workflows/*.yaml`.
- Inspects Codex prompt and schema files when relevant.
- Emits evidence-bound findings in Markdown, JSON, and SARIF 2.1.0.
- Builds structured review packets for humans or Codex.
- Explains each rule with examples, remediation, and false-positive notes.

This is intentionally not a broad agentic workflow scanner. v0 focuses on the OpenAI Codex GitHub Action provider pack.

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
codex-action-guard init --profile pr-review-readonly
codex-action-guard audit --all --format markdown
codex-action-guard audit .github/workflows/codex.yml --format sarif --output codex-action-guard.sarif
codex-action-guard diff main...HEAD --fail-on high
codex-action-guard packet --target human --changed main...HEAD
codex-action-guard explain CODX001
```

## Safe profiles

The `init` command writes a workflow, prompt, output schema, and threat-model document. It refuses to overwrite existing files unless `--force` is set, then validates generated output by running the internal audit.

Available profiles:

- `pr-review-readonly`
- `ci-failure-analysis-readonly`
- `release-notes-draft`
- `security-review-readonly`
- `label-gated-maintainer-task`

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

## Output formats

Markdown reports are designed for maintainers. JSON reports provide a stable schema with metadata, scanned files, findings, detected invocations, safe patterns, and profile suggestions. SARIF output is suitable for code scanning ingestion.

## Development

```sh
go test ./...
go run ./cmd/codex-action-guard audit --all --fail-on high
```

The repository dogfoods the scanner in CI by auditing its own GitHub Actions workflows.
