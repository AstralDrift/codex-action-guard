# CODX009: Write-capable Codex workflow lacks trusted gate

Default severity: high

Detects write-capable Codex jobs or workflows with no obvious actor, `allow-users`, maintainer label, protected environment, or manual approval gate.

Safer pattern:

- Use `allow-users`.
- Check `github.actor` or maintainer association.
- Require maintainer labels or protected environments.
- Prefer `workflow_dispatch` for privileged maintainer tasks.

False-positive notes:

The analyzer cannot see all organization policies. Treat uncertain findings as review prompts.
