# Contributing

Thanks for helping make Codex GitHub Action workflows safer.

`codex-action-guard` is an independent community project. It is not affiliated with, endorsed by, or certified by OpenAI.

## Development setup

Requirements:

- Go 1.25 or newer.
- Git.

Run the local checks:

```sh
go test ./...
go vet ./...
go run ./cmd/codex-action-guard audit --all --fail-on high
```

## Commit style

Use Conventional Commits:

- `feat: add rule for unsafe output sink`
- `fix: avoid false positive for read-only workflow`
- `docs: clarify profile threat model`
- `test: cover direct codex exec detection`
- `chore: update dependency metadata`

Use a scope when it helps, for example `fix(audit): ...` or `docs(rules): ...`.

## Pull request checklist

- The change is scoped to Codex GitHub Action workflows.
- New or changed rules include focused tests.
- Findings remain evidence-bound and avoid overclaiming exploitability.
- False-positive notes and remediation are clear.
- Markdown, JSON, and SARIF output remain stable where applicable.
- `go test ./...` passes.
- `go vet ./...` passes.

## Rule quality bar

Rules should prefer precision over volume. A good finding includes:

- The exact source, prompt boundary, Codex invocation, privilege context, or sink when applicable.
- File and line evidence when possible.
- Why the pattern matters.
- A safer pattern the maintainer can actually apply.
- False-positive notes.

Do not add rules that merely say a workflow is "using AI" or "could be risky." The project is about concrete trust-boundary, secret, permission, sandbox, and output-sink risks.

## Adding a rule

1. Add the rule documentation in `internal/guard/rules.go`.
2. Implement deterministic detection in `internal/guard/audit.go`.
3. Add tests in `internal/guard/audit_test.go`.
4. Update [docs/rules.md](docs/rules.md).
5. Verify Markdown, JSON, and SARIF output still render cleanly.

## Reporting false positives

False positives are important bugs for this project. Please include:

- The smallest workflow snippet that reproduces the finding.
- The command you ran.
- The output format used.
- Why the workflow is safe or intentionally gated.

Remove secrets, tokens, private repository names, and sensitive paths before sharing.
