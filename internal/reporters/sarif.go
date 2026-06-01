package reporters

import "github.com/AstralDrift/codex-action-guard/internal/guard"

func SARIF(report guard.Report) ([]byte, error) {
	return guard.RenderSARIF(report)
}
