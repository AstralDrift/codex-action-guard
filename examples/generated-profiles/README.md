# Generated profiles

Generate an example profile into a temporary directory:

```sh
tmpdir="$(mktemp -d)"
codex-action-guard init --profile pr-review-readonly --out "$tmpdir"
find "$tmpdir" -type f | sort
```

The command writes a workflow, prompt file, output schema, and threat-model document, then audits the generated output.
