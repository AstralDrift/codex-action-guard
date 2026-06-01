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

## SARIF

SARIF output targets SARIF 2.1.0. It includes rule metadata, locations, severity mapped to SARIF levels, and finding context in `properties`.

Example:

```sh
codex-action-guard audit --format sarif --output codex-action-guard.sarif
```

SARIF upload is not enabled by default in the generated profiles because this project avoids granting write-capable permissions unless explicitly needed.
