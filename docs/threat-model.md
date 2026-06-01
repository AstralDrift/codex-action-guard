# Threat model

`codex-action-guard` focuses on Codex running inside GitHub Actions. It does not guarantee security and does not replace GitHub security controls, branch protection, CodeQL, actionlint, zizmor, review, or least privilege.

## Trusted sources

Trusted sources are repository-controlled files from a trusted branch or protected review path, such as:

- Reviewed workflow files.
- Reviewed prompt files under `.github/codex/prompts`.
- Reviewed schema files under `.github/codex/schemas`.
- Maintainer-controlled workflow inputs.
- Protected environment approvals.

## Untrusted sources

Untrusted sources include attacker-controlled or contributor-controlled text:

- Pull request titles and bodies.
- Pull request head refs and branch names.
- Issue, discussion, and comment bodies.
- Commit messages.
- Workflow run titles, branch names, and artifacts unless produced by a trusted job.
- Prompt or schema changes from an untrusted pull request.

PR numbers and stable repository identifiers are usually not treated as untrusted prose because they do not carry instructions for Codex to follow.

## Codex prompt boundary

The prompt boundary is any place where text becomes instructions or task input for Codex:

- `with.prompt`
- `with.prompt-file`
- stdin or arguments to `codex exec`
- shell-generated prompt files consumed by Codex

Unsafe prompt boundaries appear when untrusted prose reaches Codex without a sanitizer, trusted gate, or careful framing.

## Codex invocation

A Codex invocation is either:

- `uses: openai/codex-action@...`
- a direct `codex exec` command in a workflow step

The analyzer records the workflow file, job, step, prompt boundary, sandbox, safety strategy, output file, and output schema when available.

## Privilege context

Privilege context describes what Codex shares a trust boundary with:

- `GITHUB_TOKEN` permissions.
- Model API keys.
- OIDC token availability.
- Checkout of repository-controlled code.
- Deployment, release, package, merge, label, or comment capabilities.
- Protected environment approval.

## Downstream sinks

Downstream sinks are places where Codex output can affect external state:

- shell commands
- `actions/github-script`
- `gh`
- release publishing
- deployment commands
- package publishing
- PR merge automation
- label/comment automation
- generated files consumed by later write-capable steps

## Risk patterns

Prompt-to-agent risk appears when untrusted text can instruct Codex. Prompt-to-script risk appears when free-form Codex output is consumed by shell or API automation.

Example unsafe trust boundary:

```yaml
with:
  prompt: "Do what this comment asks: ${{ github.event.comment.body }}"
```

Example safer pattern:

```yaml
with:
  prompt-file: .github/codex/prompts/review.md
  output-schema-file: .github/codex/schemas/review.schema.json
  sandbox: read-only
```

Non-example:

```yaml
with:
  prompt: "Review PR #${{ github.event.pull_request.number }}."
```

The PR number identifies a target. It does not itself carry prose instructions.
