# Review packets

Generate a packet for a human reviewer:

```sh
codex-action-guard packet --target human --output review-packet.md
```

Generate a packet intended to be pasted into Codex:

```sh
codex-action-guard packet --target codex --output codex-review-packet.md
```

Codex-targeted packets instruct the reviewer not to run arbitrary commands and not to invent vulnerabilities.
