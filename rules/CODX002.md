# CODX002: Untrusted content reaches Codex prompt in write-capable job

Default severity: high

Escalates CODX001 when the same job has write permissions, secrets, OIDC, deployment access, or later write-capable side effects.

Examples:

- Issue comment body is sent to Codex in a job with `issues: write`.
- Pull request text reaches Codex before a `gh pr merge` step.

Safer pattern:

- Split read-only Codex analysis from write-capable follow-up jobs.
- Require a trusted gate before writes.

False-positive notes:

This rule is context-sensitive. A visible trusted gate or read-only job may reduce the risk.
