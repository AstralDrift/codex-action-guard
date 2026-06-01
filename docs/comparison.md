# Comparison

`codex-action-guard` complements other security and CI tools. It is intentionally narrow.

## actionlint

actionlint validates GitHub Actions syntax and common workflow mistakes. `codex-action-guard` focuses on Codex-specific trust boundaries, prompts, model API keys, sandbox choices, and output sinks.

Use both.

## zizmor

zizmor finds GitHub Actions security issues broadly. `codex-action-guard` focuses on how Codex prompt boundaries interact with untrusted GitHub event content, secrets, permissions, and downstream automation.

Use both.

## CodeQL

CodeQL analyzes source code and selected workflow/security patterns. `codex-action-guard` analyzes Codex workflow composition and generated safe profiles.

Use both.

## Broad agentic workflow scanners

Broad scanners may support many agents, CI systems, or automation platforms. v0 of this project does not. It focuses on the OpenAI Codex GitHub Action provider pack so rules can be precise.

## AI code reviewers

AI code reviewers inspect code changes. `codex-action-guard` is a deterministic scanner for workflow safety. It does not call an LLM by default and is not a generic code reviewer.

## Generic prompt eval tools

Prompt eval tools assess model behavior. `codex-action-guard` assesses workflow trust boundaries: who controls the prompt, what privileges are present, and where Codex output goes.
