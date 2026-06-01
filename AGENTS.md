# Agent guidance

`codex-action-guard` is an independent community project. It is not affiliated with, endorsed by, or certified by OpenAI.

## Build and test

- Run `go test ./...` before submitting code changes.
- Run `go vet ./...` for static checks.
- Run `go run ./cmd/codex-action-guard audit --all --fail-on high` before changing workflows.
- When changing generated output intentionally, run `go test ./internal/guard -update` and review the golden diff.

## Coding style

- Keep the default audit path deterministic and LLM-free.
- Prefer standard library packages where practical.
- Keep rule logic conservative and evidence-bound.
- Use `gopkg.in/yaml.v3` when source positions matter.
- Preserve stable JSON fields; add new fields instead of renaming existing ones.

## Security rules

- Do not add default LLM calls to scanning, reporting, testing, or CI.
- Do not broaden v0 beyond Codex GitHub Action workflows.
- Do not claim complete security coverage.
- Avoid "this is exploitable" language unless a concrete source-to-boundary-to-sink path is shown.
- Keep vulnerable and secure fixtures paired when practical.
- Keep generated profiles safe by default.

## Review expectations

- Rule changes need tests for unsafe and safe cases.
- Fixture changes should explain the trust boundary being modeled.
- Docs should use careful language: "unsafe trust boundary" and "review required" are preferred when exploitability is not proven.
