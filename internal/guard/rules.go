package guard

import "fmt"

type RuleDoc struct {
	ID                 string
	Title              string
	DefaultSeverity    Severity
	Summary            string
	Examples           []string
	FalsePositiveNotes string
	Remediation        string
	SafePatterns       []string
	References         []string
}

var ruleDocs = []RuleDoc{
	{
		ID:              "CODX001",
		Title:           "Untrusted GitHub event content reaches Codex prompt",
		DefaultSeverity: SeverityMedium,
		Summary:         "Detects attacker-controlled GitHub event fields interpolated into Codex prompts, prompt files, stdin, or shell-generated prompt files.",
		Examples: []string{
			"with.prompt includes github.event.pull_request.body",
			"a shell step writes github.event.comment.body into a prompt file consumed by codex exec",
		},
		FalsePositiveNotes: "The finding may be acceptable when the source is strongly gated or only trusted maintainers can control the field. Keep the gate near the Codex invocation and document it.",
		Remediation:        "Use static prompt files, pass only stable identifiers such as PR numbers, fetch untrusted text through a sanitizer, or require a maintainer-controlled gate.",
		SafePatterns: []string{
			"prompt-file points at a reviewed file in .github/codex/prompts",
			"untrusted PR text is summarized by deterministic code before it reaches Codex",
		},
		References: []string{"openai/codex-action README", "GitHub Actions security hardening"},
	},
	{
		ID:                 "CODX002",
		Title:              "Untrusted content reaches Codex prompt in write-capable job",
		DefaultSeverity:    SeverityHigh,
		Summary:            "Escalates CODX001 when the same job has write permissions, secrets, OIDC, deployment access, or later write-capable side effects.",
		Examples:           []string{"pull_request_target job reads a PR body into prompt and later runs gh pr merge"},
		FalsePositiveNotes: "The rule is confidence-sensitive. If the job only has read permissions and no write sink, CODX001 should be the primary finding.",
		Remediation:        "Split read-only Codex generation from write-capable follow-up jobs and require schema validation plus a trusted gate before any write.",
		SafePatterns:       []string{"read-only Codex job uploads an artifact; separate maintainer-approved job performs writes"},
		References:         []string{"openai/codex-action Security page", "GitHub token permissions"},
	},
	{
		ID:                 "CODX003",
		Title:              "OpenAI or Codex API key exposed at job scope with repo-controlled code",
		DefaultSeverity:    SeverityHigh,
		Summary:            "Flags job-level OPENAI_API_KEY or CODEX_API_KEY when the job checks out or runs repository-controlled code.",
		Examples:           []string{"jobs.codex.env.OPENAI_API_KEY is set and the job runs actions/checkout followed by npm test"},
		FalsePositiveNotes: "Step-scoped use directly on openai/codex-action is the expected safe shape and should not trigger this rule.",
		Remediation:        "Pass secrets only to the Codex action input or the smallest possible step scope. Avoid job-level env for model API keys.",
		SafePatterns:       []string{"with.openai-api-key: ${{ secrets.OPENAI_API_KEY }} on the Codex action step"},
		References:         []string{"openai/codex-action README"},
	},
	{
		ID:                 "CODX004",
		Title:              "Codex uses danger-full-access or unsafe strategy without trusted trigger or gate",
		DefaultSeverity:    SeverityHigh,
		Summary:            "Detects dangerous sandbox or safety strategy choices, especially on untrusted triggers.",
		Examples:           []string{"with.sandbox: danger-full-access", "with.safety-strategy: unsafe", "codex exec --sandbox danger-full-access"},
		FalsePositiveNotes: "Some self-hosted or Windows workflows may require unsafe strategy. They still need trusted prompts and tight gating.",
		Remediation:        "Prefer read-only or workspace-write with privilege reduction. Add actor allowlists, maintainer labels, environments, or manual dispatch.",
		SafePatterns:       []string{"sandbox: read-only with safety-strategy: drop-sudo"},
		References:         []string{"openai/codex-action Safety Strategy"},
	},
	{
		ID:                 "CODX005",
		Title:              "Codex output feeds sensitive sink without schema validation",
		DefaultSeverity:    SeverityHigh,
		Summary:            "Detects Codex final messages or output files consumed by shell, github-script, gh, release, deploy, package publish, merge, label, or comment automation without structured validation.",
		Examples:           []string{"steps.run_codex.outputs.final-message is sent directly to actions/github-script createComment"},
		FalsePositiveNotes: "Artifact upload alone is not treated as a sensitive sink. The rule focuses on automation that changes external state.",
		Remediation:        "Require output-schema or output-schema-file, validate with jq/ajv/jsonschema, and constrain downstream commands.",
		SafePatterns:       []string{"output-schema-file plus jq validation before posting or deploying"},
		References:         []string{"SARIF 2.1.0", "openai/codex-action Outputs"},
	},
	{
		ID:                 "CODX006",
		Title:              "Privileged trigger checks out untrusted code before Codex",
		DefaultSeverity:    SeverityCritical,
		Summary:            "Detects pull_request_target or workflow_run workflows that check out PR head or workflow-run code before Codex or write-capable steps.",
		Examples:           []string{"pull_request_target plus actions/checkout ref: ${{ github.event.pull_request.head.sha }} before Codex"},
		FalsePositiveNotes: "Checking out the base branch on pull_request_target is not the same as checking out attacker-controlled head code.",
		Remediation:        "Do not checkout untrusted head code in privileged jobs. Use pull_request with read-only permissions or split into an unprivileged job.",
		SafePatterns:       []string{"pull_request_target job checks out the base ref only and gates writes by maintainer label"},
		References:         []string{"GitHub Actions pull_request_target security"},
	},
	{
		ID:                 "CODX007",
		Title:              "Codex job has broad GITHUB_TOKEN permissions",
		DefaultSeverity:    SeverityMedium,
		Summary:            "Flags missing explicit permissions, write-all, or broad write permissions on jobs invoking Codex.",
		Examples:           []string{"a Codex job omits permissions", "permissions: write-all"},
		FalsePositiveNotes: "Some workflows need narrow write scopes. The finding should name the exact write permissions seen.",
		Remediation:        "Set explicit minimal permissions at the job level. Prefer contents: read for read-only Codex jobs.",
		SafePatterns:       []string{"permissions: { contents: read, pull-requests: read }"},
		References:         []string{"GitHub Actions GITHUB_TOKEN permissions"},
	},
	{
		ID:                 "CODX008",
		Title:              "Prompt or schema file modified in same change that triggers Codex",
		DefaultSeverity:    SeverityMedium,
		Summary:            "In diff mode, detects changed prompt or schema files referenced by Codex workflows.",
		Examples:           []string{".github/codex/prompts/review.md changes with .github/workflows/codex-review.yml"},
		FalsePositiveNotes: "Changing a prompt is often legitimate. The risk is that trusted instructions or output constraints changed with workflow behavior.",
		Remediation:        "Review prompt and schema changes as code. Require maintainer approval before write-capable Codex workflows use them.",
		SafePatterns:       []string{"separate PRs for prompt/schema changes and workflow permission changes"},
		References:         []string{"Prompt injection review guidance"},
	},
	{
		ID:                 "CODX009",
		Title:              "Write-capable Codex workflow lacks trusted gate",
		DefaultSeverity:    SeverityHigh,
		Summary:            "Detects write-capable Codex jobs without actor, allow-users, maintainer label, environment, or manual approval gates.",
		Examples:           []string{"issue_comment workflow invokes Codex and comments back without actor checks"},
		FalsePositiveNotes: "The rule recognizes obvious gates but cannot prove branch protection or organization policy. Treat medium confidence as a review prompt.",
		Remediation:        "Add allow-users, actor allowlists, maintainer label checks, protected environments, or workflow_dispatch-only execution.",
		SafePatterns:       []string{"if condition requires a maintainer label and the action has allow-users set"},
		References:         []string{"openai/codex-action allow-users", "GitHub environments"},
	},
	{
		ID:                 "CODX010",
		Title:              "Codex output is posted without size limits, escaping, redaction, or schema constraints",
		DefaultSeverity:    SeverityMedium,
		Summary:            "Detects free-form Codex output posted to PR/issue comments, releases, summaries, or generated files without constraints.",
		Examples:           []string{"CODEX_FINAL_MESSAGE is used as a GitHub issue comment body with no truncation or escaping"},
		FalsePositiveNotes: "Free-form human-only artifacts are lower risk. Posting to public comments or releases needs more care.",
		Remediation:        "Use schema constraints, length limits, escaping, and secret redaction before posting output.",
		SafePatterns:       []string{"validated JSON plus explicit truncation before createComment"},
		References:         []string{"openai/codex-action output-schema-file"},
	},
}

func AllRuleDocs() []RuleDoc {
	out := make([]RuleDoc, len(ruleDocs))
	copy(out, ruleDocs)
	return out
}

func GetRuleDoc(id string) (RuleDoc, bool) {
	for _, doc := range ruleDocs {
		if doc.ID == id {
			return doc, true
		}
	}
	return RuleDoc{}, false
}

func ExplainRule(id string) (string, error) {
	doc, ok := GetRuleDoc(id)
	if !ok {
		return "", fmt.Errorf("unknown rule id %q", id)
	}
	return RenderRuleDoc(doc), nil
}
