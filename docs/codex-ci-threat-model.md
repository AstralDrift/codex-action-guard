# Codex CI threat model

`codex-action-guard` is an independent community project. It is not affiliated with, endorsed by, or certified by OpenAI.

This document describes the trust boundaries the project uses when generating and auditing Codex GitHub Action workflows.

## Assets

- OpenAI or compatible model API keys stored in GitHub Actions secrets.
- `GITHUB_TOKEN` scopes available to workflow jobs.
- Repository contents, release artifacts, and generated files.
- Trusted prompt files in `.github/codex/prompts`.
- Trusted output schemas in `.github/codex/schemas`.
- Human reviewer trust in Codex output.

## Untrusted inputs

- Pull request titles, bodies, branch names, head refs, and commit messages.
- Issue, discussion, and comment bodies.
- Workflow run display text, logs, and artifacts unless produced by a trusted job.
- Prompt and schema changes from untrusted pull requests.

## Safe defaults

- Prefer read-only Codex jobs.
- Declare explicit job-level permissions.
- Use `actions/checkout` with `persist-credentials: false`.
- Pass API keys through the Codex action input instead of job-level environment variables.
- Use `prompt-file` for static trusted instructions.
- Constrain Codex output with `output-schema-file` before downstream automation consumes it.
- Split read-only generation from write-capable follow-up jobs.
- Require `workflow_dispatch`, `allow-users`, actor checks, maintainer labels, protected environments, or manual approval for write-capable workflows.

## Review questions

- What untrusted source can influence the prompt?
- What secrets or write tokens are available in the same job?
- Can Codex read or modify repository-controlled code before or after the model call?
- Does Codex output reach shell, `gh`, `actions/github-script`, release, deploy, publish, merge, label, or comment sinks?
- Are prompt and schema changes reviewed as trusted workflow behavior changes?
