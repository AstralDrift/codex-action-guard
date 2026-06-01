package guard

import (
	"sort"
	"strings"

	"github.com/AstralDrift/codex-action-guard/internal/githubactions"
)

func addSafePatterns(report *Report, inv *codexInvocation) {
	if inv.summary.PromptFile != "" {
		addSafePattern(report, SafePattern{File: inv.summary.File, Line: inv.summary.Line, Pattern: "Codex uses a checked-in prompt-file instead of large inline prompt text."})
	}
	if inv.hasSchema {
		addSafePattern(report, SafePattern{File: inv.summary.File, Line: inv.summary.Line, Pattern: "Codex output is constrained by an output schema."})
	}
	if strings.EqualFold(inv.summary.Sandbox, "read-only") {
		addSafePattern(report, SafePattern{File: inv.summary.File, Line: inv.summary.Line, Pattern: "Codex runs with read-only sandbox settings."})
	}
	if inv.job.permissions.Explicit && !inv.job.permissions.WriteAll && len(inv.job.permissions.Writes) == 0 {
		addSafePattern(report, SafePattern{File: inv.summary.File, Line: inv.job.permissions.Line, Pattern: "Codex job declares read-only or empty GITHUB_TOKEN permissions."})
	}
	for _, step := range inv.job.steps {
		if isCheckoutStep(step) {
			if node, ok := step.with["persist-credentials"]; ok && strings.EqualFold(githubactions.Scalar(node), "false") {
				addSafePattern(report, SafePattern{File: inv.summary.File, Line: node.Line, Pattern: "actions/checkout disables persisted credentials."})
			}
		}
	}
	if inv.job.hasGate {
		addSafePattern(report, SafePattern{File: inv.summary.File, Line: inv.job.line, Pattern: "Codex job has an obvious trusted gate such as actor, label, allow-users, or environment approval."})
	}
}

func addFinding(report *Report, finding Finding) {
	if finding.Line == 0 {
		finding.Line = 1
	}
	for _, existing := range report.Findings {
		if existing.RuleID == finding.RuleID && existing.File == finding.File && existing.Line == finding.Line && existing.CodexInvocation == finding.CodexInvocation && existing.DownstreamSink == finding.DownstreamSink {
			return
		}
	}
	report.Findings = append(report.Findings, finding)
	sort.SliceStable(report.Findings, func(i, j int) bool {
		if SeverityRank(report.Findings[i].Severity) != SeverityRank(report.Findings[j].Severity) {
			return SeverityRank(report.Findings[i].Severity) > SeverityRank(report.Findings[j].Severity)
		}
		if report.Findings[i].File != report.Findings[j].File {
			return report.Findings[i].File < report.Findings[j].File
		}
		return report.Findings[i].Line < report.Findings[j].Line
	})
}

func addSafePattern(report *Report, pattern SafePattern) {
	for _, existing := range report.SafePatterns {
		if existing.File == pattern.File && existing.Line == pattern.Line && existing.Pattern == pattern.Pattern {
			return
		}
	}
	report.SafePatterns = append(report.SafePatterns, pattern)
}

func profileSuggestions(report Report) []string {
	if len(report.CodexWorkflowFiles) == 0 {
		return []string{"No Codex workflows found. Start with `codex-action-guard init --profile pr-review-readonly` for a safe read-only profile."}
	}
	suggestions := []string{}
	seen := map[string]bool{}
	for _, finding := range report.Findings {
		switch finding.RuleID {
		case "CODX007":
			seen["permissions"] = true
		case "CODX001", "CODX002":
			seen["prompt"] = true
		case "CODX005", "CODX010":
			seen["schema"] = true
		case "CODX009":
			seen["gate"] = true
		}
	}
	if seen["permissions"] {
		suggestions = append(suggestions, "Set explicit job-level permissions; read-only Codex jobs usually need contents: read and sometimes pull-requests: read.")
	}
	if seen["prompt"] {
		suggestions = append(suggestions, "Move static instructions into .github/codex/prompts and keep attacker-controlled text outside the prompt boundary unless gated.")
	}
	if seen["schema"] {
		suggestions = append(suggestions, "Use output-schema-file before feeding Codex output to comments, releases, shell, gh, or deploy automation.")
	}
	if seen["gate"] {
		suggestions = append(suggestions, "Add allow-users, actor checks, maintainer labels, protected environments, or workflow_dispatch-only entrypoints for write-capable workflows.")
	}
	return suggestions
}
