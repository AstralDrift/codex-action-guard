# Install guard workflow

Use `codex-action-guard install` to add a safe audit workflow to another repository.

```sh
codex-action-guard install
```

By default, the command writes:

```text
.github/workflows/codex-action-guard.yml
```

It refuses to overwrite an existing workflow unless `--force` is set.

## Options

- `--preset artifact|sarif`: choose the generated workflow shape. Defaults to `artifact`.
- `--out <repo>`: target repository path. Defaults to the current directory.
- `--force`: overwrite `.github/workflows/codex-action-guard.yml` when it already exists.

## Artifact preset

```sh
codex-action-guard install --preset artifact
```

The artifact preset is the safest default. It uses only `contents: read`, runs `AstralDrift/codex-action-guard@v0` with Markdown output, and uploads the report as a workflow artifact.

See [`../examples/install/codex-action-guard-artifact.yml`](../examples/install/codex-action-guard-artifact.yml).

## SARIF preset

```sh
codex-action-guard install --preset sarif
```

The SARIF preset adds `security-events: write`, emits SARIF, and uploads results to GitHub code scanning. Use it when the target repository has code scanning enabled and you want findings in the Security tab.

See [`../examples/install/codex-action-guard-sarif.yml`](../examples/install/codex-action-guard-sarif.yml).

## Generated workflow safety properties

Both presets include:

- `actions/checkout@v6` with `persist-credentials: false`.
- `AstralDrift/codex-action-guard@v0` with `fail-on: high`.
- `workflow_dispatch` for manual runs.
- Path filters for `.github/workflows/**`, `.github/codex/**`, `AGENTS.md`, and `action.yml`.
- `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true` to avoid GitHub Actions Node 20 runtime deprecation warnings for JavaScript actions.

The `@v0` action reference is a floating major tag maintained by the release process and currently points at the latest compatible v0 release.
