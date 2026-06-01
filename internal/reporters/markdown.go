package reporters

import "github.com/AstralDrift/codex-action-guard/internal/guard"

func Markdown(report guard.Report) string {
	return guard.RenderMarkdown(report)
}
