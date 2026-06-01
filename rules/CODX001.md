# CODX001: Untrusted GitHub event content reaches Codex prompt

Default severity: medium

Detects attacker-controlled GitHub event fields interpolated into `with.prompt`, prompt files, stdin or arguments to `codex exec`, or shell-generated prompt files.

Examples:

- `${{ github.event.pull_request.body }}` in `with.prompt`
- `${{ github.event.comment.body }}` written to a prompt file

Safer pattern:

- Use trusted `prompt-file` instructions.
- Pass stable identifiers such as PR numbers.
- Sanitize or gate untrusted text before it reaches Codex.

False-positive notes:

This may be acceptable when only trusted maintainers control the source and the gate is explicit near the Codex invocation.
