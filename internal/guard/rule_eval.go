package guard

import "github.com/AstralDrift/codex-action-guard/internal/githubactions"

func evaluateInvocationRules(wf workflowInfo, jobs []*jobInfo, inv *codexInvocation, report *Report) {
	untrusted := findInvocationUntrustedMatches(inv)
	if len(untrusted) > 0 {
		for _, match := range untrusted {
			f := findingFromRule("CODX001", wf, match.Line,
				match.Source,
				match.Boundary,
				invocationLabel(inv),
				inv.job.permissions.Description(),
				"",
				[]Evidence{{
					File:        match.File,
					Line:        match.Line,
					Description: match.Description,
					Snippet:     githubactions.LineSnippet(match.Lines, match.Line),
				}},
			)
			f.File = match.File
			addFinding(report, f)
		}
		if jobHasSensitiveContext(inv.job) {
			f := findingFromRule("CODX002", wf, untrusted[0].Line,
				untrusted[0].Source,
				untrusted[0].Boundary,
				invocationLabel(inv),
				sensitivePrivilegeContext(inv.job),
				"",
				[]Evidence{{
					File:        untrusted[0].File,
					Line:        untrusted[0].Line,
					Description: "The untrusted prompt source shares a job with secrets, write permissions, OIDC, or write-capable sinks.",
					Snippet:     githubactions.LineSnippet(untrusted[0].Lines, untrusted[0].Line),
				}},
			)
			f.File = untrusted[0].File
			addFinding(report, f)
		}
	}

	if unsafe, line := unsafeCodexMode(inv); unsafe != "" {
		severity := SeverityHigh
		if len(untrusted) > 0 && hasUntrustedTrigger(wf.triggers) {
			severity = SeverityCritical
		} else if inv.job.hasGate || trustedTriggerOnly(wf.triggers) {
			severity = SeverityMedium
		}
		f := findingFromRule("CODX004", wf, line, "", inv.summary.PromptBoundary, invocationLabel(inv), inv.job.permissions.Description(), "", []Evidence{
			evidence(wf, line, unsafe),
		})
		f.Severity = severity
		addFinding(report, f)
	}

	if privilegedTrigger(wf.triggers) {
		for _, step := range inv.job.steps {
			if step.index >= inv.step.index {
				break
			}
			if isUntrustedCheckout(step) {
				addFinding(report, findingFromRule("CODX006", wf, step.line,
					"pull_request_target/workflow_run untrusted checkout",
					inv.summary.PromptBoundary,
					invocationLabel(inv),
					inv.job.permissions.Description(),
					"checkout before Codex",
					[]Evidence{evidence(wf, step.line, "Privileged workflow checks out attacker-influenced code before Codex runs.")},
				))
			}
		}
	}

	sinks := downstreamSinks(inv, jobs)
	for _, sink := range sinks {
		if !inv.hasSchema && sink.Kind != "" {
			addFinding(report, findingFromRule("CODX005", wf, sink.Line,
				"",
				inv.summary.PromptBoundary,
				invocationLabel(inv),
				inv.job.permissions.Description(),
				sink.Detail,
				[]Evidence{{File: wf.relFile, Line: sink.Line, Description: "Codex output feeds a sensitive downstream sink without output-schema validation.", Snippet: sink.Snippet}},
			))
		}
		if sink.Posting && !inv.hasSchema && !sink.HasConstraint {
			addFinding(report, findingFromRule("CODX010", wf, sink.Line,
				"",
				inv.summary.PromptBoundary,
				invocationLabel(inv),
				inv.job.permissions.Description(),
				sink.Detail,
				[]Evidence{{File: wf.relFile, Line: sink.Line, Description: "Free-form Codex output appears to be posted without schema, size limit, escaping, or redaction.", Snippet: sink.Snippet}},
			))
		}
	}
}

func evaluateJobRules(wf workflowInfo, job *jobInfo, report *Report) {
	for _, secret := range job.envSecrets {
		if (secret.Key == "OPENAI_API_KEY" || secret.Key == "CODEX_API_KEY") && job.repoControlled {
			addFinding(report, findingFromRule("CODX003", wf, secret.Line,
				"job env "+secret.Key,
				"job-level environment",
				"Codex job "+job.id,
				job.permissions.Description(),
				"",
				[]Evidence{evidence(wf, secret.Line, "API key is exposed at job scope while the job checks out or runs repository-controlled code.")},
			))
		}
	}

	if job.permissions.Missing || job.permissions.WriteAll || job.permissions.HasBroadWrites() {
		f := findingFromRule("CODX007", wf, job.permissions.Line,
			"GITHUB_TOKEN permissions",
			"",
			"Codex job "+job.id,
			job.permissions.Description(),
			"",
			[]Evidence{evidence(wf, nonZero(job.permissions.Line, job.line), job.permissions.Description())},
		)
		if job.permissions.WriteAll || job.permissions.HasBroadWrites() {
			f.Severity = SeverityHigh
		}
		addFinding(report, f)
	}

	if jobWriteCapable(job) && hasUntrustedTrigger(wf.triggers) && !job.hasGate {
		addFinding(report, findingFromRule("CODX009", wf, job.line,
			"write-capable Codex job on untrusted trigger",
			"",
			"Codex job "+job.id,
			job.permissions.Description(),
			"",
			[]Evidence{evidence(wf, job.line, "The job can write or perform write-capable side effects but no obvious actor, label, allow-users, environment, or manual gate was found.")},
		))
	}
}

func evaluateDiffRules(wf workflowInfo, invocations []*codexInvocation, changed map[string]bool, report *Report) {
	if len(changed) == 0 {
		return
	}
	for _, inv := range invocations {
		for _, ref := range []string{inv.summary.PromptFile, inv.summary.OutputSchemaFile} {
			if ref == "" {
				continue
			}
			rel := githubactions.NormalizeWorkflowRef(ref)
			if changed[rel] {
				addFinding(report, findingFromRule("CODX008", wf, inv.summary.Line,
					rel,
					inv.summary.PromptBoundary,
					invocationLabel(inv),
					inv.job.permissions.Description(),
					"",
					[]Evidence{{File: rel, Description: "Referenced prompt or schema file changed in this diff."}},
				))
			}
		}
	}
}

func findingFromRule(ruleID string, wf workflowInfo, line int, source string, promptBoundary string, invocation string, privilege string, sink string, evidences []Evidence) Finding {
	doc, _ := GetRuleDoc(ruleID)
	return Finding{
		RuleID:             doc.ID,
		Title:              doc.Title,
		Severity:           doc.DefaultSeverity,
		Confidence:         ConfidenceHigh,
		File:               wf.relFile,
		Line:               line,
		Source:             source,
		PromptBoundary:     promptBoundary,
		CodexInvocation:    invocation,
		PrivilegeContext:   privilege,
		DownstreamSink:     sink,
		Evidence:           evidences,
		WhyItMatters:       doc.Summary,
		SaferPattern:       doc.Remediation,
		FalsePositiveNotes: doc.FalsePositiveNotes,
		References:         doc.References,
	}
}

func evidence(wf workflowInfo, line int, description string) Evidence {
	return Evidence{
		File:        wf.relFile,
		Line:        line,
		Description: description,
		Snippet:     githubactions.LineSnippet(wf.lines, line),
	}
}
