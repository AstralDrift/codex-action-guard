package reporters

import "github.com/AstralDrift/codex-action-guard/internal/guard"

func JSON(report guard.Report) ([]byte, error) {
	return guard.RenderJSON(report)
}
