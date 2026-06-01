# CODX006: Privileged trigger checks out untrusted code before Codex

Default severity: critical

Detects `pull_request_target` or `workflow_run` workflows that check out attacker-influenced refs before Codex or before write-capable steps.

Safer pattern:

- Do not check out PR head code in privileged jobs.
- Use `pull_request` with read-only permissions for untrusted code.
- On `pull_request_target`, check out the trusted base unless a maintainer-approved gate is present.

False-positive notes:

Checking out the base branch is not the same as checking out attacker-controlled head code.
