package rules

import "github.com/AstralDrift/codex-action-guard/internal/guard"

type Severity = guard.Severity
type Confidence = guard.Confidence

const (
	SeverityInfo     = guard.SeverityInfo
	SeverityLow      = guard.SeverityLow
	SeverityMedium   = guard.SeverityMedium
	SeverityHigh     = guard.SeverityHigh
	SeverityCritical = guard.SeverityCritical
)

func SeverityRank(severity Severity) int {
	return guard.SeverityRank(severity)
}
