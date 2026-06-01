# Usage guide

## `version`

Print build metadata:

```sh
codex-action-guard version
```

## `init`

Generate a safe profile:

```sh
codex-action-guard init --profile pr-review-readonly
```

Options:

- `--profile <name>`: required profile name.
- `--out <repo>`: output repository path. Defaults to the current directory.
- `--force`: overwrite generated files.

The command audits generated output before exiting.

## `audit`

Audit workflows in the current repository:

```sh
codex-action-guard audit --all
```

Audit one workflow:

```sh
codex-action-guard audit .github/workflows/codex.yml --format json
```

Options:

- `--format markdown|json|sarif`: default is `markdown`.
- `--output <file>`: write the report to a file.
- `--fail-on info|low|medium|high|critical|none`: exit non-zero when the threshold is met. Default is `none`.
- `--all`: include Codex prompt files, schema files, and `AGENTS.md` in the scanned-file list.

Exit codes:

- `0`: command completed and the fail threshold was not met.
- `1`: command failed.
- `2`: invalid usage.
- `3`: audit completed and the `--fail-on` threshold was met.

## `diff`

Audit Codex-relevant files changed in a git range:

```sh
codex-action-guard diff main...HEAD --fail-on high
```

The command scans changed workflow files and includes workflow context when changed prompt or schema files are referenced by Codex workflows.

## `packet`

Produce a structured review packet:

```sh
codex-action-guard packet --target human --changed main...HEAD
```

Options:

- `--target human|codex`: tune packet wording for a human reviewer or Codex.
- `--changed <rev-range>`: include changed-file context from a git diff range.
- `--output <file>`: write the packet to a file.

Packets include detected Codex invocations, prompt boundaries, privilege context, downstream sinks, findings, safe-pattern suggestions, and review questions.

## `explain`

Print rule documentation:

```sh
codex-action-guard explain CODX001
```

Rule docs include examples, remediation, safe patterns, false-positive notes, and references.
