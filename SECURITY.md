# Security policy

`codex-action-guard` is an independent community project. It is not affiliated with, endorsed by, or certified by OpenAI.

## Supported versions

The project is currently in v0. Until the first stable release, security fixes target the `main` branch and the latest published release, if any.

## Reporting a vulnerability

Please do not open a public issue for a suspected vulnerability.

Use GitHub's private vulnerability reporting or security advisory flow for this repository. If that is not available, contact a repository maintainer privately through GitHub.

Include:

- A concise description of the issue.
- Steps to reproduce.
- Impact and affected versions or commits.
- Whether the issue can expose secrets, broaden token permissions, misclassify unsafe workflows, or produce unsafe generated profiles.

Do not include live secrets, tokens, private keys, or proprietary workflow content.

## Scope

Security reports are in scope when they affect:

- Generated workflow profile safety.
- Audit rule correctness for concrete trust-boundary risks.
- Handling of workflow, prompt, schema, or repository paths.
- SARIF, Markdown, or JSON output that could mislead maintainers.
- CI or release automation for this repository.

Reports about unsupported providers or generic AI workflows are normally out of scope for v0 unless they also affect supported Codex GitHub Action behavior.

## Disclosure expectations

Maintainers will acknowledge valid reports as quickly as practical, triage impact, and coordinate a fix before public disclosure. Because this is an early open-source project, exact response times may vary.
