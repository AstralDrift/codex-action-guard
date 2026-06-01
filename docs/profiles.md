# Generated profiles

`codex-action-guard init --profile <name>` writes:

- `.github/workflows/codex-<profile>.yml`
- `.github/codex/prompts/<profile>.md`
- `.github/codex/schemas/<profile>.schema.json`
- `docs/codex-ci-threat-model.md` when it does not already exist

The command refuses to overwrite profile-specific files unless `--force` is set. The shared threat-model document is skipped when it already exists so multiple profiles can be added to the same repository.

After generation, the CLI audits the output.

## Shared safety choices

- Explicit job-level permissions.
- `actions/checkout` with `persist-credentials: false`.
- API keys passed through the Codex action input.
- Static `prompt-file` instructions.
- `output-schema-file` for structured output.
- Read-only Codex sandbox where possible.
- No default posting, merging, publishing, or deployment automation.

## pr-review-readonly

Use for pull request review artifacts.

The workflow checks out the pull request merge commit with read-only permissions, runs Codex with a trusted prompt file and JSON schema, and uploads the result as an artifact. It does not post comments or write to the repository.

## ci-failure-analysis-readonly

Use for read-only failure analysis after a CI workflow fails.

The workflow checks out the default branch, analyzes repository context, and uploads a structured artifact. It treats workflow run text, logs, branch names, commit messages, and artifacts as untrusted.

## release-notes-draft

Use for manually triggered release note drafting.

The workflow is `workflow_dispatch` only, takes maintainer-selected refs, and produces a structured draft artifact. It does not publish a release or push tags.

## security-review-readonly

Use for evidence-bound security review artifacts.

The workflow can run on pull requests or manually. It produces structured findings with severity, confidence, evidence, remediation, and false-positive notes.

## label-gated-maintainer-task

Use for a maintainer-approved task where the workflow has write-capable permissions.

The workflow uses `pull_request_target` carefully: it checks out the trusted base, requires a maintainer-style label and author association, uses a protected environment, and passes `allow-users` to the Codex action. It still uploads an artifact rather than directly applying writes.
