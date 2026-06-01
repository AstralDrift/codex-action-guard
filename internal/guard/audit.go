package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type workflowInfo struct {
	absFile  string
	relFile  string
	raw      string
	lines    []string
	root     *yaml.Node
	triggers map[string]bool
}

type jobInfo struct {
	id             string
	node           *yaml.Node
	line           int
	permissions    permissionContext
	envSecrets     []envSecret
	steps          []stepInfo
	hasCheckout    bool
	repoControlled bool
	hasGate        bool
	hasEnvironment bool
	secretUse      bool
}

type stepInfo struct {
	index int
	node  *yaml.Node
	line  int
	id    string
	name  string
	uses  string
	run   string
	with  map[string]*yaml.Node
}

type codexInvocation struct {
	summary          Invocation
	job              *jobInfo
	step             stepInfo
	promptText       string
	promptLine       int
	codexArgs        string
	outputSchemaText string
	hasSchema        bool
}

type permissionContext struct {
	Explicit bool
	Line     int
	Missing  bool
	WriteAll bool
	ReadAll  bool
	Writes   []string
	Reads    []string
	IDToken  bool
}

type envSecret struct {
	Key  string
	Line int
}

type untrustedMatch struct {
	Source string
	Line   int
	Text   string
}

type sinkInfo struct {
	Line          int
	Kind          string
	Detail        string
	Snippet       string
	Posting       bool
	HasConstraint bool
}

var codexExecPattern = regexp.MustCompile(`(?m)(^|[;&|[:space:]])codex[[:space:]]+exec\b`)

var untrustedPatterns = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"pull request title", regexp.MustCompile(`github\.event\.pull_request\.title`)},
	{"pull request body", regexp.MustCompile(`github\.event\.pull_request\.body`)},
	{"pull request head ref", regexp.MustCompile(`github\.event\.pull_request\.head\.(ref|label|sha|repo\.full_name)`)},
	{"issue title", regexp.MustCompile(`github\.event\.issue\.title`)},
	{"issue body", regexp.MustCompile(`github\.event\.issue\.body`)},
	{"comment body", regexp.MustCompile(`github\.event\.comment\.body`)},
	{"discussion content", regexp.MustCompile(`github\.event\.discussion(_comment)?\.(title|body)`)},
	{"branch name", regexp.MustCompile(`github\.(head_ref|ref_name)`)},
	{"commit message", regexp.MustCompile(`github\.event\.(head_commit\.message|commits)`)},
	{"workflow run text", regexp.MustCompile(`github\.event\.workflow_run\.(display_title|head_branch|head_sha|name)`)},
	{"artifact-derived text", regexp.MustCompile(`(?i)(download-artifact|artifact.*(body|comment|prompt|message|summary))`)},
}

func AuditPath(target string, opts AuditOptions) (Report, error) {
	if target == "" {
		target = "."
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return Report{}, err
	}
	info, err := os.Stat(absTarget)
	if err != nil {
		return Report{}, err
	}
	rootStart := absTarget
	if !info.IsDir() {
		rootStart = filepath.Dir(absTarget)
	}
	root := findRepoRoot(rootStart)
	report := NewReport(root, opts.ToolVersion)

	files, err := collectAuditFiles(root, absTarget, info, opts)
	if err != nil {
		return Report{}, err
	}
	sort.Strings(files)
	report.ScannedFiles = files

	changed := map[string]bool{}
	for _, file := range opts.ChangedFiles {
		changed[filepath.ToSlash(file)] = true
	}

	for _, rel := range files {
		if !isWorkflowFile(rel) {
			continue
		}
		if err := analyzeWorkflow(root, rel, changed, opts, &report); err != nil {
			return Report{}, err
		}
	}

	report.ProfileSuggestions = profileSuggestions(report)
	return report, nil
}

func collectAuditFiles(root string, absTarget string, info os.FileInfo, opts AuditOptions) ([]string, error) {
	if len(opts.ChangedFiles) > 0 {
		var files []string
		needsWorkflowContext := false
		for _, changed := range opts.ChangedFiles {
			rel := filepath.ToSlash(strings.TrimPrefix(filepath.Clean(changed), string(filepath.Separator)))
			if !isRelevantFile(rel) {
				continue
			}
			if isPromptOrSchemaFile(rel) {
				needsWorkflowContext = true
			}
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
				files = append(files, rel)
			}
		}
		if needsWorkflowContext {
			for _, pattern := range []string{
				filepath.Join(root, ".github", "workflows", "*.yml"),
				filepath.Join(root, ".github", "workflows", "*.yaml"),
			} {
				matches, err := filepath.Glob(pattern)
				if err != nil {
					return nil, err
				}
				for _, match := range matches {
					rel, err := filepath.Rel(root, match)
					if err != nil {
						return nil, err
					}
					files = append(files, filepath.ToSlash(rel))
				}
			}
		}
		return uniqueStrings(files), nil
	}

	if !info.IsDir() {
		rel, err := filepath.Rel(root, absTarget)
		if err != nil {
			return nil, err
		}
		return []string{filepath.ToSlash(rel)}, nil
	}

	var files []string
	for _, pattern := range []string{
		filepath.Join(root, ".github", "workflows", "*.yml"),
		filepath.Join(root, ".github", "workflows", "*.yaml"),
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			rel, err := filepath.Rel(root, match)
			if err != nil {
				return nil, err
			}
			files = append(files, filepath.ToSlash(rel))
		}
	}
	if opts.All {
		for _, pattern := range []string{
			filepath.Join(root, ".github", "codex", "prompts", "*.md"),
			filepath.Join(root, ".github", "codex", "schemas", "*.json"),
		} {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, err
			}
			for _, match := range matches {
				rel, err := filepath.Rel(root, match)
				if err != nil {
					return nil, err
				}
				files = append(files, filepath.ToSlash(rel))
			}
		}
		if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err == nil {
			files = append(files, "AGENTS.md")
		}
	}
	return uniqueStrings(files), nil
}

func analyzeWorkflow(root string, rel string, changed map[string]bool, opts AuditOptions, report *Report) error {
	abs := filepath.Join(root, filepath.FromSlash(rel))
	data, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	parsed, err := parseYAML(data)
	if err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	wf := workflowInfo{
		absFile:  abs,
		relFile:  rel,
		raw:      string(data),
		lines:    strings.Split(string(data), "\n"),
		root:     parsed,
		triggers: workflowTriggers(parsed),
	}

	workflowPermissions := parsePermissions(mapLookup(parsed, "permissions"))
	jobsNode := mapLookup(parsed, "jobs")
	var jobs []*jobInfo
	var invocations []*codexInvocation

	for _, jobPair := range mapPairs(jobsNode) {
		if jobPair.Value.Kind != yaml.MappingNode {
			continue
		}
		job := buildJobInfo(jobPair.Key.Value, jobPair.Value, workflowPermissions)
		stepsNode := mapLookup(jobPair.Value, "steps")
		if stepsNode == nil {
			jobs = append(jobs, job)
			continue
		}
		for i, stepNode := range stepsNode.Content {
			if stepNode.Kind != yaml.MappingNode {
				continue
			}
			step := buildStepInfo(i, stepNode)
			job.steps = append(job.steps, step)
			if isCheckoutStep(step) {
				job.hasCheckout = true
			}
			if stepRunsRepoControlledCode(step) {
				job.repoControlled = true
			}
			if stepHasGate(step) {
				job.hasGate = true
			}
		}
		if job.hasCheckout {
			job.repoControlled = true
		}
		if strings.Contains(nodeText(job.node), "secrets.") {
			job.secretUse = true
		}
		jobs = append(jobs, job)
		for _, step := range job.steps {
			if inv, ok := buildCodexInvocation(wf, job, step); ok {
				invocations = append(invocations, inv)
			}
		}
	}

	if len(invocations) == 0 {
		return nil
	}

	report.CodexWorkflowFiles = appendUnique(report.CodexWorkflowFiles, rel)
	for _, inv := range invocations {
		report.CodexInvocations = append(report.CodexInvocations, inv.summary)
		addSafePatterns(report, inv)
	}

	jobRuleDone := map[string]bool{}
	for _, inv := range invocations {
		evaluateInvocationRules(wf, jobs, inv, report)
		if !jobRuleDone[inv.job.id] {
			evaluateJobRules(wf, inv.job, report)
			jobRuleDone[inv.job.id] = true
		}
	}
	if opts.DiffMode {
		evaluateDiffRules(wf, invocations, changed, report)
	}
	return nil
}

func buildJobInfo(id string, node *yaml.Node, workflowPermissions permissionContext) *jobInfo {
	jobPermissions := parsePermissions(mapLookup(node, "permissions"))
	if !jobPermissions.Explicit && workflowPermissions.Explicit {
		jobPermissions = workflowPermissions
	}
	job := &jobInfo{
		id:             id,
		node:           node,
		line:           node.Line,
		permissions:    jobPermissions,
		envSecrets:     jobScopedSecrets(node),
		hasGate:        jobHasGate(node),
		hasEnvironment: mapLookup(node, "environment") != nil,
	}
	if job.hasEnvironment {
		job.hasGate = true
	}
	return job
}

func buildStepInfo(index int, node *yaml.Node) stepInfo {
	with := mapToStringNodes(mapLookup(node, "with"))
	return stepInfo{
		index: index,
		node:  node,
		line:  node.Line,
		id:    scalarValue(mapLookup(node, "id")),
		name:  scalarValue(mapLookup(node, "name")),
		uses:  scalarValue(mapLookup(node, "uses")),
		run:   scalarValue(mapLookup(node, "run")),
		with:  with,
	}
}

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
		promptText = scalarValue(node)
		promptLine = node.Line
		promptBoundary = "with.prompt"
	}
	if node, ok := step.with["prompt-file"]; ok {
		promptFile = scalarValue(node)
		promptBoundary = "prompt-file: " + promptFile
	}
	if node, ok := step.with["output-file"]; ok {
		outputFile = scalarValue(node)
	}
	if node, ok := step.with["output-schema-file"]; ok {
		outputSchema = scalarValue(node)
	}
	if node, ok := step.with["output-schema"]; ok {
		outputSchema = scalarValue(node)
	}
	if node, ok := step.with["sandbox"]; ok {
		sandbox = scalarValue(node)
	}
	if node, ok := step.with["safety-strategy"]; ok {
		safetyStrategy = scalarValue(node)
	}
	if node, ok := step.with["codex-args"]; ok {
		codexArgs = scalarValue(node)
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
		promptText:       promptText,
		promptLine:       promptLine,
		codexArgs:        codexArgs,
		outputSchemaText: outputSchema,
		hasSchema:        outputSchema != "" || strings.Contains(codexArgs, "--output-schema"),
	}, true
}

func evaluateInvocationRules(wf workflowInfo, jobs []*jobInfo, inv *codexInvocation, report *Report) {
	untrusted := findUntrustedMatches(inv.promptText, inv.promptLine, wf.raw)
	if len(untrusted) > 0 {
		for _, match := range untrusted {
			addFinding(report, findingFromRule("CODX001", wf, match.Line,
				match.Source,
				inv.summary.PromptBoundary,
				invocationLabel(inv),
				inv.job.permissions.Description(),
				"",
				[]Evidence{evidence(wf, match.Line, "Untrusted GitHub event content is interpolated into the Codex prompt boundary.")},
			))
		}
		if jobHasSensitiveContext(inv.job) {
			addFinding(report, findingFromRule("CODX002", wf, untrusted[0].Line,
				untrusted[0].Source,
				inv.summary.PromptBoundary,
				invocationLabel(inv),
				sensitivePrivilegeContext(inv.job),
				"",
				[]Evidence{evidence(wf, untrusted[0].Line, "The untrusted prompt source shares a job with secrets, write permissions, OIDC, or write-capable sinks.")},
			))
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

	if job.permissions.Missing || job.permissions.WriteAll || hasBroadWrites(job.permissions) {
		f := findingFromRule("CODX007", wf, job.permissions.Line,
			"GITHUB_TOKEN permissions",
			"",
			"Codex job "+job.id,
			job.permissions.Description(),
			"",
			[]Evidence{evidence(wf, nonZero(job.permissions.Line, job.line), job.permissions.Description())},
		)
		if job.permissions.WriteAll || hasBroadWrites(job.permissions) {
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
			rel := normalizeWorkflowRef(ref)
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
		Snippet:     lineSnippet(wf.lines, line),
	}
}

func parsePermissions(node *yaml.Node) permissionContext {
	if node == nil {
		return permissionContext{Missing: true}
	}
	ctx := permissionContext{Explicit: true, Line: node.Line}
	if node.Kind == yaml.ScalarNode {
		switch strings.ToLower(strings.TrimSpace(node.Value)) {
		case "write-all":
			ctx.WriteAll = true
		case "read-all":
			ctx.ReadAll = true
		case "{}":
			ctx.Reads = append(ctx.Reads, "none")
		}
		return ctx
	}
	for _, pair := range mapPairs(node) {
		key := pair.Key.Value
		value := strings.ToLower(strings.TrimSpace(scalarValue(pair.Value)))
		switch value {
		case "write":
			ctx.Writes = append(ctx.Writes, key)
			if key == "id-token" {
				ctx.IDToken = true
			}
		case "read":
			ctx.Reads = append(ctx.Reads, key)
		}
	}
	sort.Strings(ctx.Writes)
	sort.Strings(ctx.Reads)
	return ctx
}

func (p permissionContext) Description() string {
	switch {
	case p.Missing:
		return "missing explicit permissions; repository GITHUB_TOKEN defaults may apply"
	case p.WriteAll:
		return "permissions: write-all"
	case len(p.Writes) > 0:
		return "write permissions: " + strings.Join(p.Writes, ", ")
	case p.ReadAll:
		return "permissions: read-all"
	case len(p.Reads) > 0:
		return "read permissions: " + strings.Join(p.Reads, ", ")
	default:
		return "permissions explicitly empty or none"
	}
}

func workflowTriggers(root *yaml.Node) map[string]bool {
	triggers := map[string]bool{}
	node := mapLookup(root, "on")
	if node == nil {
		return triggers
	}
	switch node.Kind {
	case yaml.ScalarNode:
		triggers[node.Value] = true
	case yaml.SequenceNode:
		for _, item := range node.Content {
			triggers[item.Value] = true
		}
	case yaml.MappingNode:
		for _, pair := range mapPairs(node) {
			triggers[pair.Key.Value] = true
		}
	}
	return triggers
}

func jobScopedSecrets(jobNode *yaml.Node) []envSecret {
	envNode := mapLookup(jobNode, "env")
	var out []envSecret
	for _, pair := range mapPairs(envNode) {
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
	repoPatterns := []string{"go test", "npm ", "pnpm ", "yarn ", "make", "./", "bash scripts", "sh scripts", "python ", "pip install", "bundle ", "cargo ", "mvn ", "gradle "}
	for _, pattern := range repoPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func jobHasGate(jobNode *yaml.Node) bool {
	text := strings.ToLower(nodeText(jobNode))
	gates := []string{"github.actor", "allow-users", "allow-bots", "author_association", "maintainer", "labels.*.name", "contains(github.event.pull_request.labels", "environment:"}
	for _, gate := range gates {
		if strings.Contains(text, gate) {
			return true
		}
	}
	return mapLookup(jobNode, "environment") != nil
}

func stepHasGate(step stepInfo) bool {
	text := strings.ToLower(nodeText(step.node))
	return strings.Contains(text, "allow-users") || strings.Contains(text, "github.actor") || strings.Contains(text, "maintainer") || strings.Contains(text, "labels")
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
			line = firstLineOf(raw, matched, fallbackLine)
		}
		matches = append(matches, untrustedMatch{
			Source: candidate.name,
			Line:   line,
			Text:   matched,
		})
	}
	return matches
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
	for _, trigger := range []string{"pull_request", "pull_request_target", "issue_comment", "issues", "discussion", "discussion_comment", "workflow_run"} {
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
		switch trigger {
		case "workflow_dispatch", "schedule":
		default:
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
		ref = strings.ToLower(scalarValue(node))
	}
	untrustedRefs := []string{"github.event.pull_request.head", "github.head_ref", "refs/pull/", "github.event.workflow_run.head_sha", "github.event.workflow_run.head_branch"}
	for _, pattern := range untrustedRefs {
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

func hasBroadWrites(permissions permissionContext) bool {
	if permissions.WriteAll {
		return true
	}
	broad := map[string]bool{
		"actions":         true,
		"attestations":    true,
		"checks":          true,
		"contents":        true,
		"deployments":     true,
		"id-token":        true,
		"packages":        true,
		"pages":           true,
		"security-events": true,
	}
	for _, scope := range permissions.Writes {
		if broad[scope] {
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
			if !mentionsCodexOutput(step, inv) {
				continue
			}
			if sink := classifySink(step, inv.summary.OutputFile); sink.Kind != "" {
				sinks = append(sinks, sink)
			}
		}
	}
	return sinks
}

func mentionsCodexOutput(step stepInfo, inv *codexInvocation) bool {
	text := strings.ToLower(nodeText(step.node))
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
	text := strings.ToLower(nodeText(step.node))
	uses := strings.ToLower(step.uses)
	run := strings.ToLower(step.run)
	line := step.line
	detail := stepLabel(step)
	snippet := strings.TrimSpace(firstNonEmpty(step.uses, step.run, step.name))

	sensitive := map[string]string{
		"actions/github-script":           "github-script automation",
		"gh pr":                           "GitHub CLI PR automation",
		"gh issue":                        "GitHub CLI issue automation",
		"gh release":                      "GitHub CLI release automation",
		"git push":                        "git push",
		"npm publish":                     "package publish",
		"pypi-publish":                    "package publish",
		"docker push":                     "container publish",
		"kubectl":                         "deployment command",
		"deploy":                          "deployment action or command",
		"createcomment":                   "GitHub comment automation",
		"issues.createcomment":            "GitHub issue comment automation",
		"pulls.merge":                     "GitHub merge automation",
		"issues.addlabels":                "GitHub label automation",
		"softprops/action-gh-release":     "GitHub release automation",
		"peter-evans/create-pull-request": "pull request write automation",
	}
	for pattern, kind := range sensitive {
		if strings.Contains(uses, pattern) || strings.Contains(run, pattern) || strings.Contains(text, pattern) {
			return sinkInfo{Line: line, Kind: kind, Detail: detail + ": " + kind, Snippet: snippet, Posting: isPostingSink(text), HasConstraint: hasOutputConstraint(text)}
		}
	}
	if strings.Contains(text, "github_step_summary") || strings.Contains(text, "$github_step_summary") {
		return sinkInfo{Line: line, Kind: "summary", Detail: detail + ": job summary posting", Snippet: snippet, Posting: true, HasConstraint: hasOutputConstraint(text)}
	}
	return sinkInfo{}
}

func isPostingSink(text string) bool {
	posting := []string{"comment", "createcomment", "github_step_summary", "release", "body:", "body =", "pr comment", "issue comment"}
	for _, pattern := range posting {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func hasOutputConstraint(text string) bool {
	constraints := []string{"output-schema", "jq", "ajv", "jsonschema", "head -c", "truncate", "wc -c", "substring", "substr", "slice(", "escape", "redact", "::add-mask::", "mask"}
	for _, pattern := range constraints {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

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
			if node, ok := step.with["persist-credentials"]; ok && strings.EqualFold(scalarValue(node), "false") {
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

func findRepoRoot(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			return start
		}
		dir = next
	}
}

func isWorkflowFile(rel string) bool {
	rel = filepath.ToSlash(rel)
	return strings.HasPrefix(rel, ".github/workflows/") && (strings.HasSuffix(rel, ".yml") || strings.HasSuffix(rel, ".yaml"))
}

func isRelevantFile(rel string) bool {
	rel = filepath.ToSlash(rel)
	if isWorkflowFile(rel) {
		return true
	}
	return strings.HasPrefix(rel, ".github/codex/prompts/") || strings.HasPrefix(rel, ".github/codex/schemas/") || rel == "AGENTS.md"
}

func isPromptOrSchemaFile(rel string) bool {
	rel = filepath.ToSlash(rel)
	return strings.HasPrefix(rel, ".github/codex/prompts/") || strings.HasPrefix(rel, ".github/codex/schemas/")
}

func normalizeWorkflowRef(ref string) string {
	ref = strings.TrimSpace(strings.Trim(ref, `"'`))
	ref = strings.TrimPrefix(ref, "./")
	return filepath.ToSlash(ref)
}

func extractFlagValue(text string, flag string) string {
	fields := strings.Fields(strings.NewReplacer("[", " ", "]", " ", ",", " ", `"`, " ", "'", " ").Replace(text))
	for i, field := range fields {
		if field == flag && i+1 < len(fields) {
			return fields[i+1]
		}
		if strings.HasPrefix(field, flag+"=") {
			return strings.TrimPrefix(field, flag+"=")
		}
	}
	return flag
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func stepLabel(step stepInfo) string {
	if step.name != "" {
		return step.name
	}
	if step.id != "" {
		return step.id
	}
	if step.uses != "" {
		return step.uses
	}
	return fmt.Sprintf("step %d", step.index+1)
}

func invocationLabel(inv *codexInvocation) string {
	return fmt.Sprintf("%s job %s at %s:%d", inv.summary.Kind, inv.job.id, inv.summary.File, inv.summary.Line)
}

func nonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 1
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
