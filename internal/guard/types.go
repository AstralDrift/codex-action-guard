package guard

import (
	"strings"
	"time"
)

const (
	ToolName    = "codex-action-guard"
	RuleVersion = "v0"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type Confidence string

const (
	ConfidenceLow    Confidence = "low"
	ConfidenceMedium Confidence = "medium"
	ConfidenceHigh   Confidence = "high"
)

type Metadata struct {
	Tool        string    `json:"tool"`
	Version     string    `json:"version"`
	RuleVersion string    `json:"rule_version"`
	GeneratedAt time.Time `json:"generated_at"`
}

type Report struct {
	Metadata           Metadata      `json:"metadata"`
	Root               string        `json:"root"`
	ScannedFiles       []string      `json:"scanned_files"`
	CodexWorkflowFiles []string      `json:"codex_workflow_files"`
	CodexInvocations   []Invocation  `json:"codex_invocations"`
	Findings           []Finding     `json:"findings"`
	SafePatterns       []SafePattern `json:"safe_patterns"`
	ProfileSuggestions []string      `json:"profile_suggestions"`
}

type Invocation struct {
	File             string `json:"file"`
	Line             int    `json:"line,omitempty"`
	Job              string `json:"job"`
	Step             string `json:"step,omitempty"`
	Kind             string `json:"kind"`
	PromptBoundary   string `json:"prompt_boundary,omitempty"`
	PromptFile       string `json:"prompt_file,omitempty"`
	OutputFile       string `json:"output_file,omitempty"`
	OutputSchemaFile string `json:"output_schema_file,omitempty"`
	Sandbox          string `json:"sandbox,omitempty"`
	SafetyStrategy   string `json:"safety_strategy,omitempty"`
	PrivilegeContext string `json:"privilege_context"`
}

type Finding struct {
	RuleID             string     `json:"rule_id"`
	Title              string     `json:"title"`
	Severity           Severity   `json:"severity"`
	Confidence         Confidence `json:"confidence"`
	File               string     `json:"file"`
	Line               int        `json:"line,omitempty"`
	Source             string     `json:"source,omitempty"`
	PromptBoundary     string     `json:"prompt_boundary,omitempty"`
	CodexInvocation    string     `json:"codex_invocation,omitempty"`
	PrivilegeContext   string     `json:"privilege_context"`
	DownstreamSink     string     `json:"downstream_sink,omitempty"`
	Evidence           []Evidence `json:"evidence"`
	WhyItMatters       string     `json:"why_it_matters"`
	SaferPattern       string     `json:"safer_pattern"`
	FalsePositiveNotes string     `json:"false_positive_notes"`
	References         []string   `json:"references"`
}

type Evidence struct {
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Description string `json:"description"`
	Snippet     string `json:"snippet,omitempty"`
}

type SafePattern struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Pattern string `json:"pattern"`
}

type AuditOptions struct {
	All          bool
	DiffMode     bool
	ChangedFiles []string
	ToolVersion  string
}

func NewReport(root string, version string) Report {
	if version == "" {
		version = "dev"
	}
	return Report{
		Metadata: Metadata{
			Tool:        ToolName,
			Version:     version,
			RuleVersion: RuleVersion,
			GeneratedAt: time.Now().UTC(),
		},
		Root:               root,
		ScannedFiles:       []string{},
		CodexWorkflowFiles: []string{},
		CodexInvocations:   []Invocation{},
		Findings:           []Finding{},
		SafePatterns:       []SafePattern{},
		ProfileSuggestions: []string{},
	}
}

func (r Report) HasCodexWorkflows() bool {
	return len(r.CodexWorkflowFiles) > 0
}

func (r Report) MeetsThreshold(threshold Severity) bool {
	if threshold == "none" {
		return false
	}
	for _, finding := range r.Findings {
		if SeverityRank(finding.Severity) >= SeverityRank(threshold) {
			return true
		}
	}
	return false
}

func ParseSeverity(value string) (Severity, bool) {
	switch Severity(strings.ToLower(strings.TrimSpace(value))) {
	case SeverityInfo:
		return SeverityInfo, true
	case SeverityLow:
		return SeverityLow, true
	case SeverityMedium:
		return SeverityMedium, true
	case SeverityHigh:
		return SeverityHigh, true
	case SeverityCritical:
		return SeverityCritical, true
	case "none":
		return Severity("none"), true
	default:
		return "", false
	}
}

func SeverityRank(severity Severity) int {
	switch severity {
	case SeverityInfo:
		return 0
	case SeverityLow:
		return 1
	case SeverityMedium:
		return 2
	case SeverityHigh:
		return 3
	case SeverityCritical:
		return 4
	default:
		return -1
	}
}

func SeverityGroups() []Severity {
	return []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
}

func SARIFLevel(severity Severity) string {
	switch severity {
	case SeverityCritical, SeverityHigh:
		return "error"
	case SeverityMedium:
		return "warning"
	case SeverityLow:
		return "note"
	default:
		return "none"
	}
}
