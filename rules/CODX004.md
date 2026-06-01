# CODX004: Unsafe Codex sandbox or safety strategy

Default severity: high

Detects `sandbox: danger-full-access`, unsafe safety strategy, or equivalent direct `codex exec` flags without a trusted trigger or gate.

Safer pattern:

- Prefer `sandbox: read-only` for generated public-repo workflows.
- Use `safety-strategy: drop-sudo` where supported.
- Add actor allowlists, labels, environments, or manual dispatch for privileged workflows.

False-positive notes:

Some self-hosted or Windows workflows may require weaker isolation. They still need trusted prompts and narrow permissions.
