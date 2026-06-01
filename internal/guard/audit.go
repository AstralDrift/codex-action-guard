package guard

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AstralDrift/codex-action-guard/internal/githubactions"
)

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
	scanRoot := root
	if len(opts.ChangedFiles) == 0 {
		if info.IsDir() {
			scanRoot = absTarget
		} else if inferred, ok := inferWorkflowRoot(absTarget); ok {
			scanRoot = inferred
		}
	} else if info.IsDir() && absTarget != root {
		scanRoot = absTarget
	}
	report := NewReport(scanRoot, opts.ToolVersion)

	files, err := collectAuditFiles(scanRoot, absTarget, info, opts)
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
		if err := analyzeWorkflow(scanRoot, rel, changed, opts, &report); err != nil {
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
	parsed, err := githubactions.ParseWorkflow(rel, data)
	if err != nil {
		return err
	}
	wf := workflowInfo{
		repoRoot: root,
		absFile:  abs,
		relFile:  rel,
		raw:      string(data),
		lines:    strings.Split(string(data), "\n"),
		doc:      parsed.Root,
		triggers: githubactions.TriggerSet(parsed.Triggers),
	}

	workflowPermissions := githubactions.ParsePermissions(githubactions.Lookup(parsed.Root, "permissions"))
	var jobs []*jobInfo
	var invocations []*codexInvocation

	for _, parsedJob := range parsed.Jobs {
		job := buildJobInfo(parsedJob, workflowPermissions)
		for _, parsedStep := range parsedJob.Steps {
			step := buildStepInfo(parsedStep)
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
		if strings.Contains(githubactions.NodeText(job.node), "secrets.") {
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

func buildJobInfo(parsed githubactions.Job, workflowPermissions permissionContext) *jobInfo {
	jobPermissions := parsed.Permissions
	if !jobPermissions.Explicit && workflowPermissions.Explicit {
		jobPermissions = workflowPermissions
	}
	job := &jobInfo{
		id:             parsed.ID,
		node:           parsed.Node,
		line:           parsed.Line,
		permissions:    jobPermissions,
		envSecrets:     jobScopedSecrets(parsed.Node),
		hasGate:        jobHasGate(parsed.Node),
		hasEnvironment: githubactions.Lookup(parsed.Node, "environment") != nil,
	}
	if job.hasEnvironment {
		job.hasGate = true
	}
	return job
}

func buildStepInfo(parsed githubactions.Step) stepInfo {
	return stepInfo{
		index: parsed.Index,
		node:  parsed.Node,
		line:  parsed.Line,
		id:    parsed.ID,
		name:  parsed.Name,
		uses:  parsed.Uses,
		run:   parsed.Run,
		with:  parsed.With,
	}
}
