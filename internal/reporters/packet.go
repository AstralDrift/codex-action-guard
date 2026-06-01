package reporters

import "github.com/AstralDrift/codex-action-guard/internal/guard"

func Packet(report guard.Report, target string, changed []string) string {
	return guard.RenderPacket(report, target, changed)
}
