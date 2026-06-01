# Generated guard workflows

These files show the workflows produced by `codex-action-guard install`.

- [`codex-action-guard-artifact.yml`](codex-action-guard-artifact.yml): default read-only Markdown artifact preset.
- [`codex-action-guard-sarif.yml`](codex-action-guard-sarif.yml): GitHub code scanning SARIF preset.

Regenerate them with:

```sh
codex-action-guard install --preset artifact --out <repo>
codex-action-guard install --preset sarif --out <repo> --force
```
