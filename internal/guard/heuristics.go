package guard

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AstralDrift/codex-action-guard/internal/githubactions"
	"gopkg.in/yaml.v3"
)

func jobScopedSecrets(jobNode *yaml.Node) []envSecret {
	envNode := githubactions.Lookup(jobNode, "env")
	var out []envSecret
	for _, pair := range githubactions.Pairs(envNode) {
		key := pair.Key.Value
		if key == "OPENAI_API_KEY" || key == "CODEX_API_KEY" {
			out = append(out, envSecret{Key: key, Line: pair.Key.Line})
		}
	}
	return out
}

func isCodexAction(step stepInfo) bool {
	uses := strings.ToLower(step.uses)
	return strings.Contains(uses, "openai/codex-action")
}

func isCheckoutStep(step stepInfo) bool {
	return strings.Contains(strings.ToLower(step.uses), "actions/checkout")
}

func stepRunsRepoControlledCode(step stepInfo) bool {
	text := strings.ToLower(step.run)
	if text == "" {
		return false
	}
	for _, pattern := range repoControlledRunPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func jobHasGate(jobNode *yaml.Node) bool {
	text := strings.ToLower(githubactions.NodeText(jobNode))
	for _, gate := range gatePatterns {
		if strings.Contains(text, gate) {
			return true
		}
	}
	return githubactions.Lookup(jobNode, "environment") != nil
}

func stepHasGate(step stepInfo) bool {
	text := strings.ToLower(githubactions.NodeText(step.node))
	for _, gate := range stepGatePatterns {
		if strings.Contains(text, gate) {
			return true
		}
	}
	return false
}

func findUntrustedMatches(text string, fallbackLine int, raw string) []untrustedMatch {
	if text == "" {
		return nil
	}
	var matches []untrustedMatch
	for _, candidate := range untrustedPatterns {
		loc := candidate.pattern.FindStringIndex(text)
		if loc == nil {
			continue
		}
		matched := text[loc[0]:loc[1]]
		line := fallbackLine
		if raw != "" {
			line = githubactions.FirstLineOf(raw, matched, fallbackLine)
		}
		matches = append(matches, untrustedMatch{
			Source: candidate.name,
			Line:   line,
			Text:   matched,
		})
	}
	return matches
}

func findInvocationUntrustedMatches(inv *codexInvocation) []untrustedMatch {
	var out []untrustedMatch
	for _, source := range inv.promptSources {
		raw := source.Raw
		if raw == "" {
			raw = source.Text
		}
		matches := findUntrustedMatches(source.Text, source.FallbackLine, raw)
		for _, match := range matches {
			match.File = source.File
			match.Lines = source.Lines
			match.Boundary = source.Boundary
			match.Description = source.Description
			out = append(out, match)
		}
	}
	return out
}

func promptFileSource(wf workflowInfo, promptFile string) (promptSource, bool) {
	rel := githubactions.NormalizeWorkflowRef(promptFile)
	if rel == "" || strings.Contains(rel, "${{") {
		return promptSource{}, false
	}
	path := filepath.Join(wf.repoRoot, filepath.FromSlash(rel))
	data, err := os.ReadFile(path)
	if err != nil {
		return promptSource{}, false
	}
	text := string(data)
	return promptSource{
		File:         rel,
		Text:         text,
		Raw:          text,
		Lines:        strings.Split(text, "\n"),
		FallbackLine: 1,
		Boundary:     "prompt-file: " + promptFile,
		Description:  "Untrusted GitHub event content appears in a prompt file consumed by Codex.",
	}, true
}

func stepWritesPromptFile(step stepInfo, promptFile string) bool {
	if step.run == "" {
		return false
	}
	run := strings.ToLower(step.run)
	promptFile = strings.ToLower(githubactions.NormalizeWorkflowRef(promptFile))
	if promptFile == "" {
		return false
	}
	base := strings.ToLower(filepath.Base(promptFile))
	if !strings.Contains(run, promptFile) && !strings.Contains(run, base) {
		return false
	}
	if len(findUntrustedMatches(step.run, step.line, step.run)) == 0 {
		return false
	}
	for _, marker := range promptFileWriteMarkers {
		if strings.Contains(run, marker) {
			return true
		}
	}
	return false
}

func unsafeCodexMode(inv *codexInvocation) (string, int) {
	unsafeText := strings.ToLower(strings.Join([]string{
		inv.summary.Sandbox,
		inv.summary.SafetyStrategy,
		inv.codexArgs,
		inv.step.run,
	}, "\n"))
	if strings.Contains(unsafeText, "danger-full-access") {
		return "Codex is configured with danger-full-access.", inv.summary.Line
	}
	if strings.EqualFold(strings.TrimSpace(inv.summary.SafetyStrategy), "unsafe") {
		return "Codex is configured with safety-strategy: unsafe.", inv.summary.Line
	}
	if strings.Contains(unsafeText, "safety-strategy: unsafe") || strings.Contains(unsafeText, "safety-strategy unsafe") || strings.Contains(unsafeText, "with.safety-strategy: unsafe") {
		return "Codex is configured with an unsafe safety strategy.", inv.summary.Line
	}
	if strings.Contains(unsafeText, "full-auto") || strings.Contains(unsafeText, "--ask-for-approval never") {
		return "Codex appears to run in a fully automated unsafe approval mode.", inv.summary.Line
	}
	return "", 0
}

func hasUntrustedTrigger(triggers map[string]bool) bool {
	for _, trigger := range untrustedTriggerNames {
		if triggers[trigger] {
			return true
		}
	}
	return false
}

func privilegedTrigger(triggers map[string]bool) bool {
	return triggers["pull_request_target"] || triggers["workflow_run"]
}

func trustedTriggerOnly(triggers map[string]bool) bool {
	if len(triggers) == 0 {
		return false
	}
	for trigger := range triggers {
		if !trustedOnlyTriggerNames[trigger] {
			return false
		}
	}
	return true
}

func isUntrustedCheckout(step stepInfo) bool {
	if !isCheckoutStep(step) {
		return false
	}
	ref := ""
	if node, ok := step.with["ref"]; ok {
		ref = strings.ToLower(githubactions.Scalar(node))
	}
	for _, pattern := range untrustedCheckoutRefs {
		if strings.Contains(ref, pattern) {
			return true
		}
	}
	return false
}

func jobHasSensitiveContext(job *jobInfo) bool {
	return job.secretUse || job.permissions.IDToken || job.permissions.WriteAll || len(job.permissions.Writes) > 0 || jobWriteCapable(job)
}

func sensitivePrivilegeContext(job *jobInfo) string {
	parts := []string{job.permissions.Description()}
	if job.secretUse {
		parts = append(parts, "uses secrets")
	}
	if job.permissions.IDToken {
		parts = append(parts, "OIDC token available")
	}
	if jobWriteCapable(job) {
		parts = append(parts, "write-capable side effects present")
	}
	return strings.Join(parts, "; ")
}

func jobWriteCapable(job *jobInfo) bool {
	if job.permissions.WriteAll || len(job.permissions.Writes) > 0 {
		return true
	}
	for _, step := range job.steps {
		if sink := classifySink(step, ""); sink.Kind != "" && sink.Kind != "summary" {
			return true
		}
	}
	return false
}

func downstreamSinks(inv *codexInvocation, jobs []*jobInfo) []sinkInfo {
	var sinks []sinkInfo
	for _, job := range jobs {
		for _, step := range job.steps {
			if job.id == inv.job.id && step.index <= inv.step.index {
				continue
			}
			if job.id != inv.job.id && !strings.Contains(strings.ToLower(githubactions.NodeText(step.node)), "needs."+strings.ToLower(inv.job.id)+".outputs") {
				continue
			}
			if !mentionsCodexOutput(step, inv) {
				continue
			}
			if sink := classifySink(step, inv.summary.OutputFile); sink.Kind != "" {
				sinks = append(sinks, sink)
			} else if step.run != "" {
				sinks = append(sinks, sinkInfo{
					Line:          step.line,
					Kind:          "shell command",
					Detail:        stepLabel(step) + ": shell command consumes Codex output",
					Snippet:       strings.TrimSpace(step.run),
					Posting:       isPostingSink(strings.ToLower(step.run)),
					HasConstraint: hasOutputConstraint(strings.ToLower(step.run)),
				})
			}
		}
	}
	return sinks
}

func mentionsCodexOutput(step stepInfo, inv *codexInvocation) bool {
	text := strings.ToLower(githubactions.NodeText(step.node))
	if inv.step.id != "" && strings.Contains(text, "steps."+strings.ToLower(inv.step.id)+".outputs.final-message") {
		return true
	}
	if inv.summary.OutputFile != "" {
		output := strings.ToLower(inv.summary.OutputFile)
		if strings.Contains(text, output) || strings.Contains(text, strings.ToLower(filepath.Base(output))) {
			return true
		}
	}
	if strings.Contains(text, "needs."+strings.ToLower(inv.job.id)+".outputs") {
		return true
	}
	return strings.Contains(text, "codex_final_message") || strings.Contains(text, "final-message") || strings.Contains(text, "codex-output")
}

func classifySink(step stepInfo, outputFile string) sinkInfo {
	text := strings.ToLower(githubactions.NodeText(step.node))
	uses := strings.ToLower(step.uses)
	run := strings.ToLower(step.run)
	line := step.line
	detail := stepLabel(step)
	snippet := strings.TrimSpace(firstNonEmpty(step.uses, step.run, step.name))

	for _, sink := range sensitiveSinkPatterns {
		if strings.Contains(uses, sink.pattern) || strings.Contains(run, sink.pattern) || strings.Contains(text, sink.pattern) {
			return sinkInfo{Line: line, Kind: sink.kind, Detail: detail + ": " + sink.kind, Snippet: snippet, Posting: isPostingSink(text), HasConstraint: hasOutputConstraint(text)}
		}
	}
	if strings.Contains(text, "github_step_summary") || strings.Contains(text, "$github_step_summary") {
		return sinkInfo{Line: line, Kind: "summary", Detail: detail + ": job summary posting", Snippet: snippet, Posting: true, HasConstraint: hasOutputConstraint(text)}
	}
	return sinkInfo{}
}

func isPostingSink(text string) bool {
	for _, pattern := range postingSinkPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func hasOutputConstraint(text string) bool {
	for _, pattern := range outputConstraintPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}
