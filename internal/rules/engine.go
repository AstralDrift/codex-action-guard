package rules

import "github.com/AstralDrift/codex-action-guard/internal/guard"

type Engine struct {
	Version string
}

func (e Engine) AuditPath(path string, all bool) (guard.Report, error) {
	return guard.AuditPath(path, guard.AuditOptions{All: all, ToolVersion: e.Version})
}
