# Safe patterns

These patterns are not a security guarantee. They are practical defaults for safer Codex GitHub Action workflows.

## Read-only Codex job

Use explicit read-only permissions:

```yaml
permissions:
  contents: read
  pull-requests: read
```

Run Codex with a trusted prompt file, output schema, and read-only sandbox.

## Split read and write jobs

Let Codex produce a structured artifact in one job. Perform comments, labels, releases, or deployments in a separate trusted and gated job after validation.

## Step-scoped API key handling

Pass the model API key directly to the Codex action:

```yaml
with:
  openai-api-key: ${{ secrets.OPENAI_API_KEY }}
```

Avoid `OPENAI_API_KEY` or `CODEX_API_KEY` at job scope when repository-controlled code can run.

## Checkout without persisted credentials

Use:

```yaml
with:
  persist-credentials: false
```

This reduces accidental token exposure to later repository-controlled steps.

## Trusted prompt file

Use `prompt-file` for static instructions reviewed on a trusted branch. Avoid writing raw PR, issue, or comment bodies into prompt files.

## Schema-constrained output

Use `output-schema-file` before any downstream automation consumes Codex output. Validate again with deterministic tools such as `jq` when posting or acting on the output.

## Maintainer label gate

Write-capable workflows should require a maintainer-controlled label, actor allowlist, protected environment, or manual dispatch.

## Actor allowlist

Use `allow-users` for Codex action invocations that should only be triggered by trusted maintainers.

## Environment approval

Use protected environments for write-capable follow-up jobs.

## Size limits and escaping

Before posting Codex output to a comment, summary, release, or generated file:

- validate structure
- truncate to a known maximum size
- escape markdown or shell-sensitive content
- redact secrets

## Avoid direct interpolation

Avoid:

```yaml
prompt: "${{ github.event.comment.body }}"
```

Prefer:

```yaml
prompt-file: .github/codex/prompts/review.md
```
