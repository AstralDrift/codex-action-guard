# ci-failure-analysis-readonly

Read-only CI failure analysis profile.

Safety properties:

- reads repository and Actions context.
- treats workflow run text, logs, branch names, commit messages, and artifacts as untrusted.
- uploads structured analysis as an artifact.
- does not rerun, comment, or deploy.
