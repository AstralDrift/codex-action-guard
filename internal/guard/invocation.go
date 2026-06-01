package guard

import (
	"strings"

	"github.com/AstralDrift/codex-action-guard/internal/githubactions"
)

func buildCodexInvocation(wf workflowInfo, job *jobInfo, step stepInfo) (*codexInvocation, bool) {
	kind := ""
	if isCodexAction(step) {
		kind = "openai/codex-action"
	} else if codexExecPattern.MatchString(step.run) {
		kind = "codex exec"
	}
	if kind == "" {
		return nil, false
	}

	promptText := ""
	promptLine := step.line
	promptBoundary := "Codex invocation"
	promptFile := ""
	outputFile := ""
	outputSchema := ""
	sandbox := ""
	safetyStrategy := ""
	codexArgs := ""

	if node, ok := step.with["prompt"]; ok {
		promptText = githubactions.Scalar(node)
		promptLine = node.Line
		promptBoundary = "with.prompt"
	}
	if node, ok := step.with["prompt-file"]; ok {
		promptFile = githubactions.Scalar(node)
		promptBoundary = "prompt-file: " + promptFile
	}
	if node, ok := step.with["output-file"]; ok {
		outputFile = githubactions.Scalar(node)
	}
	if node, ok := step.with["output-schema-file"]; ok {
		outputSchema = githubactions.Scalar(node)
	}
	if node, ok := step.with["output-schema"]; ok {
		outputSchema = githubactions.Scalar(node)
	}
	if node, ok := step.with["sandbox"]; ok {
		sandbox = githubactions.Scalar(node)
	}
	if node, ok := step.with["safety-strategy"]; ok {
		safetyStrategy = githubactions.Scalar(node)
	}
	if node, ok := step.with["codex-args"]; ok {
		codexArgs = githubactions.Scalar(node)
		if outputSchema == "" && strings.Contains(codexArgs, "--output-schema") {
			outputSchema = extractFlagValue(codexArgs, "--output-schema")
		}
		if sandbox == "" && strings.Contains(codexArgs, "danger-full-access") {
			sandbox = "danger-full-access"
		}
	}
	if promptText == "" && step.run != "" {
		promptText = step.run
		promptBoundary = "codex exec run script"
		promptLine = step.line
	}
	if step.name != "" {
		promptBoundary = step.name + " (" + promptBoundary + ")"
	}

	promptSources := []promptSource{}
	if promptText != "" {
		promptSources = append(promptSources, promptSource{
			File:         wf.relFile,
			Text:         promptText,
			Raw:          wf.raw,
			Lines:        wf.lines,
			FallbackLine: promptLine,
			Boundary:     promptBoundary,
			Description:  "Untrusted GitHub event content is interpolated into the Codex prompt boundary.",
		})
	}
	if promptFile != "" {
		if source, ok := promptFileSource(wf, promptFile); ok {
			promptSources = append(promptSources, source)
		}
		for _, previous := range job.steps {
			if previous.index >= step.index {
				break
			}
			if stepWritesPromptFile(previous, promptFile) {
				promptSources = append(promptSources, promptSource{
					File:         wf.relFile,
					Text:         previous.run,
					Raw:          wf.raw,
					Lines:        wf.lines,
					FallbackLine: previous.line,
					Boundary:     stepLabel(previous) + " (shell-generated prompt-file: " + promptFile + ")",
					Description:  "A shell step appears to write untrusted content into a prompt file consumed by Codex.",
				})
			}
		}
	}

	inv := Invocation{
		File:             wf.relFile,
		Line:             step.line,
		Job:              job.id,
		Step:             stepLabel(step),
		Kind:             kind,
		PromptBoundary:   promptBoundary,
		PromptFile:       promptFile,
		OutputFile:       outputFile,
		OutputSchemaFile: outputSchema,
		Sandbox:          sandbox,
		SafetyStrategy:   safetyStrategy,
		PrivilegeContext: job.permissions.Description(),
	}
	return &codexInvocation{
		summary:          inv,
		job:              job,
		step:             step,
		promptSources:    promptSources,
		codexArgs:        codexArgs,
		outputSchemaText: outputSchema,
		hasSchema:        outputSchema != "" || strings.Contains(codexArgs, "--output-schema"),
	}, true
}
