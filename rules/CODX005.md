# CODX005: Codex output feeds sensitive sink without schema validation

Default severity: high

Detects Codex `final-message`, `output-file`, artifacts, or generated files consumed by shell, `actions/github-script`, `gh`, release, deploy, publish, merge, label, or comment automation without structured validation.

Safer pattern:

- Use `output-schema` or `output-schema-file`.
- Validate with deterministic tooling before acting on output.
- Keep write-capable sinks in a separate gated job.

False-positive notes:

Artifact upload alone is not considered a sensitive sink.
