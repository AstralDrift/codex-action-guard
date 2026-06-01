# CODX003: API key exposed at job scope with repo-controlled code

Default severity: high

Flags job-level `OPENAI_API_KEY` or `CODEX_API_KEY` when the job checks out code, installs dependencies, runs scripts, runs tests, or otherwise executes repository-controlled code.

Safer pattern:

- Pass secrets through `with.openai-api-key` on the Codex action step.
- Avoid model API keys in job-level `env`.

False-positive notes:

Step-scoped secret usage directly on `openai/codex-action` should not trigger this rule.
