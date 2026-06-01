package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func RenderMarkdown(report Report) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# codex-action-guard audit report\n\n")
	fmt.Fprintf(&buf, "- Tool: `%s %s`\n", report.Metadata.Tool, report.Metadata.Version)
	fmt.Fprintf(&buf, "- Rule version: `%s`\n", report.Metadata.RuleVersion)
	fmt.Fprintf(&buf, "- Root: `%s`\n", report.Root)
	fmt.Fprintf(&buf, "- Scanned files: `%d`\n", len(report.ScannedFiles))
	fmt.Fprintf(&buf, "- Codex workflow files: `%d`\n\n", len(report.CodexWorkflowFiles))

	counts := severityCounts(report.Findings)
	fmt.Fprintf(&buf, "## Summary counts\n\n")
	for _, severity := range SeverityGroups() {
		fmt.Fprintf(&buf, "- %s: %d\n", severity, counts[severity])
	}
	fmt.Fprintf(&buf, "\n")

	if len(report.CodexWorkflowFiles) == 0 {
		fmt.Fprintf(&buf, "## Not applicable / no Codex workflows found\n\n")
		fmt.Fprintf(&buf, "No `openai/codex-action` or direct `codex exec` invocations were found in scanned workflow files.\n\n")
		writeSuggestions(&buf, report.ProfileSuggestions)
		return buf.String()
	}

	if len(report.Findings) == 0 {
		fmt.Fprintf(&buf, "## Findings\n\n")
		fmt.Fprintf(&buf, "No findings.\n\n")
	} else {
		fmt.Fprintf(&buf, "## Findings\n\n")
		for _, severity := range SeverityGroups() {
			group := findingsBySeverity(report.Findings, severity)
			if len(group) == 0 {
				continue
			}
			fmt.Fprintf(&buf, "### %s\n\n", strings.Title(string(severity)))
			for _, finding := range group {
				fmt.Fprintf(&buf, "#### %s: %s\n\n", finding.RuleID, finding.Title)
				fmt.Fprintf(&buf, "- Severity: `%s`\n", finding.Severity)
				fmt.Fprintf(&buf, "- Confidence: `%s`\n", finding.Confidence)
				fmt.Fprintf(&buf, "- Location: `%s:%d`\n", finding.File, finding.Line)
				if finding.Source != "" {
					fmt.Fprintf(&buf, "- Source: %s\n", finding.Source)
				}
				if finding.PromptBoundary != "" {
					fmt.Fprintf(&buf, "- Prompt boundary: %s\n", finding.PromptBoundary)
				}
				if finding.CodexInvocation != "" {
					fmt.Fprintf(&buf, "- Codex invocation: %s\n", finding.CodexInvocation)
				}
				fmt.Fprintf(&buf, "- Privilege context: %s\n", finding.PrivilegeContext)
				if finding.DownstreamSink != "" {
					fmt.Fprintf(&buf, "- Downstream sink: %s\n", finding.DownstreamSink)
				}
				fmt.Fprintf(&buf, "\nEvidence:\n")
				for _, ev := range finding.Evidence {
					if ev.Line > 0 {
						fmt.Fprintf(&buf, "- `%s:%d`: %s", ev.File, ev.Line, ev.Description)
					} else {
						fmt.Fprintf(&buf, "- `%s`: %s", ev.File, ev.Description)
					}
					if ev.Snippet != "" {
						fmt.Fprintf(&buf, " `%s`", ev.Snippet)
					}
					fmt.Fprintf(&buf, "\n")
				}
				fmt.Fprintf(&buf, "\nWhy it matters: %s\n\n", finding.WhyItMatters)
				fmt.Fprintf(&buf, "Safer pattern: %s\n\n", finding.SaferPattern)
				fmt.Fprintf(&buf, "False-positive notes: %s\n\n", finding.FalsePositiveNotes)
			}
		}
	}

	fmt.Fprintf(&buf, "## Safe patterns found\n\n")
	if len(report.SafePatterns) == 0 {
		fmt.Fprintf(&buf, "No safe patterns were detected.\n\n")
	} else {
		for _, pattern := range report.SafePatterns {
			if pattern.Line > 0 {
				fmt.Fprintf(&buf, "- `%s:%d`: %s\n", pattern.File, pattern.Line, pattern.Pattern)
			} else {
				fmt.Fprintf(&buf, "- `%s`: %s\n", pattern.File, pattern.Pattern)
			}
		}
		fmt.Fprintf(&buf, "\n")
	}
	writeSuggestions(&buf, report.ProfileSuggestions)
	return buf.String()
}

func RenderJSON(report Report) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

func RenderSARIF(report Report) ([]byte, error) {
	rules := make([]map[string]any, 0, len(ruleDocs))
	for _, doc := range ruleDocs {
		rules = append(rules, map[string]any{
			"id":   doc.ID,
			"name": doc.Title,
			"shortDescription": map[string]any{
				"text": doc.Title,
			},
			"fullDescription": map[string]any{
				"text": doc.Summary,
			},
			"help": map[string]any{
				"text": RenderRuleDoc(doc),
			},
			"properties": map[string]any{
				"defaultSeverity": doc.DefaultSeverity,
			},
		})
	}

	results := make([]map[string]any, 0, len(report.Findings))
	for _, finding := range report.Findings {
		result := map[string]any{
			"ruleId": finding.RuleID,
			"level":  SARIFLevel(finding.Severity),
			"message": map[string]any{
				"text": fmt.Sprintf("%s: %s", finding.RuleID, finding.Title),
			},
			"locations": []map[string]any{
				{
					"physicalLocation": map[string]any{
						"artifactLocation": map[string]any{
							"uri": finding.File,
						},
						"region": map[string]any{
							"startLine": nonZero(finding.Line, 1),
						},
					},
				},
			},
			"properties": map[string]any{
				"severity":             finding.Severity,
				"confidence":           finding.Confidence,
				"source":               finding.Source,
				"prompt_boundary":      finding.PromptBoundary,
				"codex_invocation":     finding.CodexInvocation,
				"privilege_context":    finding.PrivilegeContext,
				"downstream_sink":      finding.DownstreamSink,
				"why_it_matters":       finding.WhyItMatters,
				"safer_pattern":        finding.SaferPattern,
				"false_positive_notes": finding.FalsePositiveNotes,
				"references":           finding.References,
			},
		}
		results = append(results, result)
	}

	driver := map[string]any{
		"name":  ToolName,
		"rules": rules,
	}
	if regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+`).MatchString(report.Metadata.Version) {
		driver["semanticVersion"] = report.Metadata.Version
	} else {
		driver["version"] = report.Metadata.Version
	}

	sarif := map[string]any{
		"version": "2.1.0",
		"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
		"runs": []map[string]any{
			{
				"tool": map[string]any{
					"driver": driver,
				},
				"results": results,
			},
		},
	}
	return json.MarshalIndent(sarif, "", "  ")
}

func RenderRuleDoc(doc RuleDoc) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# %s: %s\n\n", doc.ID, doc.Title)
	fmt.Fprintf(&buf, "Default severity: `%s`\n\n", doc.DefaultSeverity)
	fmt.Fprintf(&buf, "%s\n\n", doc.Summary)
	if len(doc.Examples) > 0 {
		fmt.Fprintf(&buf, "## Examples\n\n")
		for _, example := range doc.Examples {
			fmt.Fprintf(&buf, "- %s\n", example)
		}
		fmt.Fprintf(&buf, "\n")
	}
	fmt.Fprintf(&buf, "## Remediation\n\n%s\n\n", doc.Remediation)
	if len(doc.SafePatterns) > 0 {
		fmt.Fprintf(&buf, "## Safe patterns\n\n")
		for _, pattern := range doc.SafePatterns {
			fmt.Fprintf(&buf, "- %s\n", pattern)
		}
		fmt.Fprintf(&buf, "\n")
	}
	fmt.Fprintf(&buf, "## False-positive notes\n\n%s\n\n", doc.FalsePositiveNotes)
	if len(doc.References) > 0 {
		fmt.Fprintf(&buf, "## References\n\n")
		for _, ref := range doc.References {
			fmt.Fprintf(&buf, "- %s\n", ref)
		}
	}
	return buf.String()
}

func RenderPacket(report Report, target string, changed []string) string {
	if target == "" {
		target = "human"
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Codex workflow review packet\n\n")
	fmt.Fprintf(&buf, "Target reviewer: `%s`\n\n", target)
	if target == "codex" {
		fmt.Fprintf(&buf, "You are reviewing Codex GitHub Action workflow safety. Do not run arbitrary commands. Do not invent vulnerabilities; only report evidence-bound issues from the packet. Use unsafe trust boundary language unless the evidence shows a concrete source-to-boundary-to-sink path.\n\n")
		fmt.Fprintf(&buf, "Expected output format:\n\n")
		fmt.Fprintf(&buf, "1. Summary\n")
		fmt.Fprintf(&buf, "2. Evidence-bound issues\n")
		fmt.Fprintf(&buf, "3. False-positive or uncertainty notes\n")
		fmt.Fprintf(&buf, "4. Recommended safer patterns\n\n")
	}
	fmt.Fprintf(&buf, "## Changed workflow summary\n\n")
	if len(changed) == 0 {
		fmt.Fprintf(&buf, "- No changed-file range was supplied. Packet is based on the current scanned workflow state.\n")
	} else {
		for _, file := range changed {
			if isRelevantFile(file) {
				fmt.Fprintf(&buf, "- `%s`\n", file)
			}
		}
	}
	fmt.Fprintf(&buf, "\n## Detected Codex invocations\n\n")
	if len(report.CodexInvocations) == 0 {
		fmt.Fprintf(&buf, "No Codex invocations were detected.\n\n")
	} else {
		for _, inv := range report.CodexInvocations {
			fmt.Fprintf(&buf, "- `%s:%d` job `%s`: %s", inv.File, inv.Line, inv.Job, inv.Kind)
			if inv.PromptBoundary != "" {
				fmt.Fprintf(&buf, "; prompt boundary: %s", inv.PromptBoundary)
			}
			if inv.PromptFile != "" {
				fmt.Fprintf(&buf, "; prompt file: `%s`", inv.PromptFile)
			}
			if inv.OutputSchemaFile != "" {
				fmt.Fprintf(&buf, "; output schema: `%s`", inv.OutputSchemaFile)
			}
			fmt.Fprintf(&buf, "; privilege: %s\n", inv.PrivilegeContext)
		}
		fmt.Fprintf(&buf, "\n")
	}

	fmt.Fprintf(&buf, "## Trust-boundary notes\n\n")
	writePacketFindingList(&buf, report, []string{"CODX001", "CODX002", "CODX006", "CODX008", "CODX009"})
	fmt.Fprintf(&buf, "## Output and downstream sinks\n\n")
	writePacketFindingList(&buf, report, []string{"CODX005", "CODX010"})
	fmt.Fprintf(&buf, "## Permission and secret context\n\n")
	writePacketFindingList(&buf, report, []string{"CODX003", "CODX004", "CODX007"})

	fmt.Fprintf(&buf, "## Finding summary\n\n")
	counts := severityCounts(report.Findings)
	for _, severity := range SeverityGroups() {
		fmt.Fprintf(&buf, "- %s: %d\n", severity, counts[severity])
	}
	fmt.Fprintf(&buf, "\n")
	writeSuggestions(&buf, report.ProfileSuggestions)

	fmt.Fprintf(&buf, "## Review questions\n\n")
	questions := []string{
		"Can every untrusted PR, issue, comment, branch, commit, and artifact field be traced before it reaches a Codex prompt boundary?",
		"Does the Codex job run with the narrowest possible GITHUB_TOKEN permissions?",
		"Are model API keys scoped to the Codex action step rather than the whole job?",
		"If Codex output is consumed by automation, is it constrained by an output schema and validated before shell, gh, github-script, release, deploy, merge, label, or comment sinks?",
		"Do write-capable Codex workflows require an actor allowlist, allow-users, maintainer label, protected environment, or manual approval?",
		"Are prompt and schema changes reviewed as trusted workflow behavior changes?",
	}
	for _, question := range questions {
		fmt.Fprintf(&buf, "- %s\n", question)
	}
	return buf.String()
}

func writePacketFindingList(buf *bytes.Buffer, report Report, ruleIDs []string) {
	wanted := map[string]bool{}
	for _, id := range ruleIDs {
		wanted[id] = true
	}
	var rows []Finding
	for _, finding := range report.Findings {
		if wanted[finding.RuleID] {
			rows = append(rows, finding)
		}
	}
	if len(rows) == 0 {
		fmt.Fprintf(buf, "No findings in this category.\n\n")
		return
	}
	for _, finding := range rows {
		fmt.Fprintf(buf, "- `%s` `%s:%d`: %s. Safer pattern: %s\n", finding.RuleID, finding.File, finding.Line, finding.Title, finding.SaferPattern)
	}
	fmt.Fprintf(buf, "\n")
}

func writeSuggestions(buf *bytes.Buffer, suggestions []string) {
	if len(suggestions) == 0 {
		return
	}
	fmt.Fprintf(buf, "## Profile suggestions\n\n")
	for _, suggestion := range suggestions {
		fmt.Fprintf(buf, "- %s\n", suggestion)
	}
	fmt.Fprintf(buf, "\n")
}

func severityCounts(findings []Finding) map[Severity]int {
	counts := map[Severity]int{}
	for _, finding := range findings {
		counts[finding.Severity]++
	}
	return counts
}

func findingsBySeverity(findings []Finding, severity Severity) []Finding {
	var out []Finding
	for _, finding := range findings {
		if finding.Severity == severity {
			out = append(out, finding)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Line < out[j].Line
	})
	return out
}
