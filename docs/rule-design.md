# Rule design

Rules should help maintainers understand unsafe trust boundaries without overclaiming.

## Principles

- Prefer precision over finding count.
- Explain source, boundary, invocation, privilege context, and sink when available.
- Report file and line evidence.
- Use confidence to reflect uncertainty.
- Include safer patterns and false-positive notes.
- Keep rules deterministic and LLM-free.

## Evidence model

Each finding should answer:

- What source is involved?
- Where does it cross the Codex prompt boundary?
- What Codex invocation receives it?
- What token, secret, OIDC, sandbox, or write context exists?
- Does output reach a downstream sink?

## Language

Use "unsafe trust boundary" or "review required" unless a concrete source-to-boundary-to-sink path is visible. Avoid "vulnerable" as a default adjective.

## Testing

Rule tests should include:

- an unsafe example
- a safe paired example when practical
- evidence file and line expectations for high-risk rules
- fixture coverage when the behavior is representative
