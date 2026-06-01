# CODX010: Codex output is posted without constraints

Default severity: medium

Detects free-form Codex output posted to PR or issue comments, releases, summaries, or generated files without size limits, escaping, redaction, or schema constraints.

Safer pattern:

- Constrain output with a schema.
- Truncate to a known maximum size.
- Escape markdown or shell-sensitive content.
- Redact secrets before posting.

False-positive notes:

Human-only artifacts are lower risk than public comments, releases, or automation inputs.
