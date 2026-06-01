package codex

import "strings"

func IsCodexAction(uses string) bool {
	return strings.Contains(strings.ToLower(uses), "openai/codex-action")
}

func IsDirectCodexExec(run string) bool {
	return strings.Contains(strings.ToLower(run), "codex exec")
}
