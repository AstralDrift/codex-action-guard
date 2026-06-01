# Report formats

`codex-action-guard audit` and `codex-action-guard diff` support `markdown`, `json`, and `sarif`.

## Markdown

Markdown is the default. It is designed for maintainers and code reviews.

It includes:

- Tool metadata.
- Summary counts by severity.
- Findings grouped by severity.
- Evidence snippets.
- Safe patterns found.
- Profile suggestions.
- A clean "no Codex workflows found" section when the scan is not applicable.

## JSON

JSON is intended for automation and downstream tooling.

The report schema is published at [`../schemas/report.schema.json`](../schemas/report.schema.json).

Top-level fields:

- `metadata`: tool name, version, rule version, and generation time.
- `root`: repository root.
- `scanned_files`: files included in the audit.
- `codex_workflow_files`: workflow files with Codex invocations.
- `codex_invocations`: detected Codex action or direct `codex exec` calls.
- `findings`: evidence-bound rule findings.
- `safe_patterns`: safe patterns found during analysis.
- `profile_suggestions`: remediation-oriented suggestions.

Each finding includes:

- `rule_id`
- `title`
- `severity`
- `confidence`
- `file`
- `line`
- `source`
- `prompt_boundary`
- `codex_invocation`
- `privilege_context`
- `downstream_sink`
- `evidence`
- `why_it_matters`
- `safer_pattern`
- `false_positive_notes`
- `references`

The JSON schema is stable for v0 consumers, but additional fields may be added before v1.

## Rule metadata

Rule metadata can be exported without running an audit:

```sh
codex-action-guard rules --format json --output codex-action-guard-rules.json
```

The rules export is deterministic and does not include a timestamp. Its schema is published at [`../schemas/rules.schema.json`](../schemas/rules.schema.json).

Top-level fields:

- `metadata`: tool name, version, and rule version.
- `rules`: rule ID, title, default severity, summary, examples, false-positive notes, remediation, safe patterns, and references.

## SARIF

SARIF output targets SARIF 2.1.0. It includes rule metadata, locations, severity mapped to SARIF levels, and finding context in `properties`.

Example:

```sh
codex-action-guard audit --format sarif --output codex-action-guard.sarif
```

SARIF upload is not enabled by default in the generated profiles because this project avoids granting write-capable permissions unless explicitly needed.
