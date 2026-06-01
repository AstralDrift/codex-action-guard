# Finding examples

Run:

```sh
codex-action-guard audit fixtures/vulnerable --all --format markdown
```

Expected rule coverage includes:

- `CODX001` for untrusted event text reaching prompts.
- `CODX003` for job-scoped API keys with repository-controlled code.
- `CODX005` for Codex output reaching shell or GitHub API sinks without schema validation.
- `CODX006` for privileged PR-head checkout.
- `CODX010` for free-form output posted directly.
