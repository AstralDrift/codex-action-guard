package guard

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
	Version string     `json:"version"`
}

type sarifRun struct {
	Results []sarifResult `json:"results"`
	Tool    sarifTool     `json:"tool"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name            string      `json:"name"`
	Rules           []sarifRule `json:"rules"`
	SemanticVersion string      `json:"semanticVersion,omitempty"`
	Version         string      `json:"version,omitempty"`
}

type sarifRule struct {
	FullDescription  sarifText           `json:"fullDescription"`
	Help             sarifText           `json:"help"`
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Properties       sarifRuleProperties `json:"properties"`
	ShortDescription sarifText           `json:"shortDescription"`
}

type sarifRuleProperties struct {
	DefaultSeverity Severity `json:"defaultSeverity"`
}

type sarifResult struct {
	Level      string                `json:"level"`
	Locations  []sarifLocation       `json:"locations"`
	Message    sarifText             `json:"message"`
	Properties sarifResultProperties `json:"properties"`
	RuleID     string                `json:"ruleId"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

type sarifText struct {
	Text string `json:"text"`
}

type sarifResultProperties struct {
	CodexInvocation    string     `json:"codex_invocation"`
	Confidence         Confidence `json:"confidence"`
	DownstreamSink     string     `json:"downstream_sink"`
	FalsePositiveNotes string     `json:"false_positive_notes"`
	PrivilegeContext   string     `json:"privilege_context"`
	PromptBoundary     string     `json:"prompt_boundary"`
	References         []string   `json:"references"`
	SaferPattern       string     `json:"safer_pattern"`
	Severity           Severity   `json:"severity"`
	Source             string     `json:"source"`
	WhyItMatters       string     `json:"why_it_matters"`
}

func RenderSARIF(report Report) ([]byte, error) {
	rules := make([]sarifRule, 0, len(ruleDocs))
	for _, doc := range ruleDocs {
		rules = append(rules, sarifRule{
			FullDescription:  sarifText{Text: doc.Summary},
			Help:             sarifText{Text: RenderRuleDoc(doc)},
			ID:               doc.ID,
			Name:             doc.Title,
			Properties:       sarifRuleProperties{DefaultSeverity: doc.DefaultSeverity},
			ShortDescription: sarifText{Text: doc.Title},
		})
	}

	results := make([]sarifResult, 0, len(report.Findings))
	for _, finding := range report.Findings {
		results = append(results, sarifResult{
			Level: SARIFLevel(finding.Severity),
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: finding.File},
						Region:           sarifRegion{StartLine: nonZero(finding.Line, 1)},
					},
				},
			},
			Message: sarifText{Text: fmt.Sprintf("%s: %s", finding.RuleID, finding.Title)},
			Properties: sarifResultProperties{
				CodexInvocation:    finding.CodexInvocation,
				Confidence:         finding.Confidence,
				DownstreamSink:     finding.DownstreamSink,
				FalsePositiveNotes: finding.FalsePositiveNotes,
				PrivilegeContext:   finding.PrivilegeContext,
				PromptBoundary:     finding.PromptBoundary,
				References:         finding.References,
				SaferPattern:       finding.SaferPattern,
				Severity:           finding.Severity,
				Source:             finding.Source,
				WhyItMatters:       finding.WhyItMatters,
			},
			RuleID: finding.RuleID,
		})
	}

	driver := sarifDriver{
		Name:  ToolName,
		Rules: rules,
	}
	if regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+`).MatchString(report.Metadata.Version) {
		driver.SemanticVersion = report.Metadata.Version
	} else {
		driver.Version = report.Metadata.Version
	}

	return json.MarshalIndent(sarifLog{
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs:    []sarifRun{{Results: results, Tool: sarifTool{Driver: driver}}},
		Version: "2.1.0",
	}, "", "  ")
}
