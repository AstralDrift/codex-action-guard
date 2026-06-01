# Rule reference

Use `codex-action-guard explain <RULE_ID>` for the most detailed local rule documentation.

## Severity model

Rules use `info`, `low`, `medium`, `high`, and `critical`. Severity is adjusted when the analyzer can see stronger context, such as untrusted triggers, dangerous sandbox settings, write permissions, secrets, OIDC, or downstream automation.

## Rules

| Rule | Default | Summary |
| --- | --- | --- |
| `CODX001` | medium | Untrusted GitHub event content reaches a Codex prompt boundary. |
| `CODX002` | high | `CODX001` in a job with secrets, write permissions, OIDC, deployment access, or write-capable sinks. |
| `CODX003` | high | `OPENAI_API_KEY` or `CODEX_API_KEY` is exposed at job scope while repository-controlled code can run. |
| `CODX004` | high | Codex uses `danger-full-access` or an unsafe strategy without a trusted trigger or gate. |
| `CODX005` | high | Codex output feeds shell, `github-script`, `gh`, release, deploy, publish, merge, label, or comment automation without schema validation. |
| `CODX006` | critical | `pull_request_target` or `workflow_run` checks out attacker-influenced code before Codex or write-capable steps. |
| `CODX007` | medium | A Codex job has missing, broad, or write-capable `GITHUB_TOKEN` permissions. |
| `CODX008` | medium | A prompt or schema file referenced by a Codex workflow changed in the same diff. |
| `CODX009` | high | A write-capable Codex workflow lacks an obvious trusted gate. |
| `CODX010` | medium | Free-form Codex output is posted without size limits, escaping, redaction, or schema constraints. |

## Trusted gates recognized in v0

The analyzer recognizes obvious gates including:

- `workflow_dispatch`.
- `allow-users`, `allow-bots`, and `allow-bot-users` on the Codex action.
- `github.actor` checks.
- Maintainer label or author-association checks.
- Protected `environment` usage.

It cannot prove all organization-level policies, branch protections, or custom authorization scripts. Those cases may still produce review-required findings.

## Rule authoring expectations

New rules should:

- Identify a concrete trust boundary or unsafe composition pattern.
- Provide file and line evidence.
- Avoid claiming exploitability without a source-to-boundary-to-sink path.
- Include remediation and false-positive notes.
- Include focused tests for unsafe and safe examples.
