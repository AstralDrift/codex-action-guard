# Codex CI threat model

codex-action-guard is an independent community project. It is not affiliated with, endorsed by, or certified by OpenAI.

This document records the trust boundaries used by generated Codex GitHub Action profiles.

## Assets

- OpenAI or compatible model API keys stored in GitHub Actions secrets.
- GITHUB_TOKEN scopes available to the workflow job.
- Repository contents and release artifacts.
- Prompt files and output schemas in .github/codex.
- Human reviewer trust in Codex output.

## Untrusted inputs

- Pull request titles, bodies, branch names, commit messages, and head refs.
- Issue, discussion, and comment bodies.
- Workflow run titles, logs, and artifacts unless produced by a trusted job.
- Any file changed by an untrusted pull request, including prompt and schema files.

## Safe defaults

- Codex jobs should declare explicit minimal permissions.
- Read-only profiles should use contents: read and avoid write-capable follow-up steps.
- actions/checkout should set persist-credentials: false unless a later trusted step truly needs credentials.
- API keys should be passed through the Codex action input, not job-level env.
- Static prompt-file usage is preferred for trusted instructions.
- Codex output should be schema constrained before any automation consumes it.
- Write-capable workflows need a trusted gate such as workflow_dispatch, allow-users, actor checks, maintainer labels, or protected environments.

## Review checklist

- What untrusted source can influence the prompt?
- What secrets or tokens are available in the same job?
- Can Codex read or modify repository-controlled code before or after the model call?
- Does any free-form Codex output reach shell, gh, github-script, release, deploy, package publish, merge, label, or comment sinks?
- Are prompt and schema changes reviewed as workflow behavior changes?
