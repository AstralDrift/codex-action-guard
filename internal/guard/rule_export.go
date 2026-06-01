package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type RuleCatalogExport struct {
	Metadata RuleCatalogMetadata `json:"metadata"`
	Rules    []RuleExport        `json:"rules"`
}

type RuleCatalogMetadata struct {
	Tool        string `json:"tool"`
	Version     string `json:"version"`
	RuleVersion string `json:"rule_version"`
}

type RuleExport struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	DefaultSeverity    Severity `json:"default_severity"`
	Summary            string   `json:"summary"`
	Examples           []string `json:"examples"`
	FalsePositiveNotes string   `json:"false_positive_notes"`
	Remediation        string   `json:"remediation"`
	SafePatterns       []string `json:"safe_patterns"`
	References         []string `json:"references"`
}

func NewRuleCatalogExport(version string) RuleCatalogExport {
	if version == "" {
		version = "dev"
	}
	rules := make([]RuleExport, 0, len(ruleDocs))
	for _, doc := range ruleDocs {
		rules = append(rules, RuleExport{
			ID:                 doc.ID,
			Title:              doc.Title,
			DefaultSeverity:    doc.DefaultSeverity,
			Summary:            doc.Summary,
			Examples:           append([]string{}, doc.Examples...),
			FalsePositiveNotes: doc.FalsePositiveNotes,
			Remediation:        doc.Remediation,
			SafePatterns:       append([]string{}, doc.SafePatterns...),
			References:         append([]string{}, doc.References...),
		})
	}
	return RuleCatalogExport{
		Metadata: RuleCatalogMetadata{
			Tool:        ToolName,
			Version:     version,
			RuleVersion: RuleVersion,
		},
		Rules: rules,
	}
}

func RenderRulesJSON(version string) ([]byte, error) {
	return json.MarshalIndent(NewRuleCatalogExport(version), "", "  ")
}

func RenderRulesMarkdown(version string) string {
	catalog := NewRuleCatalogExport(version)
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# %s rule catalog\n\n", ToolName)
	fmt.Fprintf(&buf, "- Tool: `%s %s`\n", catalog.Metadata.Tool, catalog.Metadata.Version)
	fmt.Fprintf(&buf, "- Rule version: `%s`\n\n", catalog.Metadata.RuleVersion)
	fmt.Fprintf(&buf, "| Rule | Default | Summary |\n")
	fmt.Fprintf(&buf, "| --- | --- | --- |\n")
	for _, rule := range catalog.Rules {
		fmt.Fprintf(&buf, "| `%s` | %s | %s |\n", rule.ID, rule.DefaultSeverity, rule.Summary)
	}
	return buf.String()
}
