# CODX008: Prompt or schema file modified with workflow behavior

Default severity: medium

In diff mode, detects prompt or schema files referenced by Codex workflows changing in the same review.

Safer pattern:

- Review prompt and schema files as trusted workflow behavior.
- Avoid combining prompt/schema changes with permission or sink changes in the same PR.

False-positive notes:

Prompt and schema changes are often legitimate. The rule asks for review because trusted agent instructions or output constraints changed.
