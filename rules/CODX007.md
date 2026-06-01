# CODX007: Codex job has broad GITHUB_TOKEN permissions

Default severity: medium

Flags missing explicit permissions, `write-all`, or broad write permissions on jobs invoking Codex. Severity increases when broad writes combine with untrusted prompts, repo-controlled code, secrets, or downstream writes.

Safer pattern:

- Set explicit job-level permissions.
- Prefer `contents: read` and only add narrow read scopes when needed.

False-positive notes:

Some workflows need narrow write scopes. The finding should name the exact write permissions seen.
