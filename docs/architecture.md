# Architecture

`codex-action-guard` is a deterministic Go CLI for generating and auditing Codex GitHub Action workflows. It does not call an LLM during audit.

## Design goals

- Be safe by default.
- Prefer deterministic static analysis over cleverness.
- Keep findings evidence-bound.
- Make false positives easy to understand and report.
- Keep v0 scoped to `openai/codex-action` and direct `codex exec` in GitHub Actions.

## Package layout

- `cmd/codex-action-guard`: CLI entrypoint.
- `internal/cli`: command parsing, report emission, git diff integration, and exit codes.
- `internal/githubactions`: GitHub Actions parsing helpers, expression classification, permission summaries, and diff path filtering.
- `internal/guard`: YAML parsing, workflow modeling, rule evaluation, rule docs, and report rendering.
- `internal/providers/codex`: OpenAI Codex GitHub Action provider facade for detection, profiles, prompts, schemas, and rules.
- `internal/profiles`: generated workflow profiles, prompt files, schemas, and threat-model text.
- `internal/reporters`: report rendering facade for Markdown, JSON, SARIF, and review packets.
- `internal/rules`: rule catalog, severity, finding, and engine facade.
- `internal/taint`: small source/sink/path types used to keep trust-boundary vocabulary explicit.

## Audit data flow

1. Resolve the target path and repository root.
2. Collect relevant workflow files from `.github/workflows`.
3. Optionally include `.github/codex/prompts`, `.github/codex/schemas`, and `AGENTS.md` in the scanned file list.
4. Parse workflows with `gopkg.in/yaml.v3` so source line numbers are available.
5. Build a workflow model: triggers, jobs, permissions, environment, steps, Codex invocations, gates, checkouts, secrets, and potential sinks.
6. Evaluate rules against that model.
7. Render Markdown, JSON, or SARIF.

## Finding philosophy

The project avoids vague "AI is risky" findings. A useful finding should describe a concrete unsafe trust boundary or review requirement:

- Source: where attacker-controlled or untrusted text may enter.
- Prompt boundary: where text reaches Codex instructions or input.
- Codex invocation: the action or direct `codex exec` call.
- Privilege context: token permissions, secrets, OIDC, or job side effects.
- Downstream sink: shell, `gh`, `actions/github-script`, release, deploy, publish, merge, label, or comment automation.

When the analyzer cannot prove a full source-to-boundary-to-sink path, it should say "unsafe trust boundary" or "review required" rather than claiming exploitability.

## Diff mode

`diff` asks git for changed files in a rev range and then limits analysis to Codex-relevant files. When prompt or schema files change, the audit includes workflow context so `CODX008` can detect referenced trusted instructions or output constraints changing in the same review.

## Limitations

- The analyzer is static and conservative.
- Shell parsing is intentionally shallow in v0.
- Repository policies, branch protection, and organization-level approval rules are not fully knowable from a workflow file.
- Some gates are recognized by obvious expressions such as actor checks, maintainer labels, `allow-users`, and protected environments.
