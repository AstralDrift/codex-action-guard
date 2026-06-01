package guard

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestVulnerableFixtureCoversInitialRules(t *testing.T) {
	report, err := AuditPath(filepath.Join("..", "..", "fixtures", "vulnerable"), AuditOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"CODX001", "CODX002", "CODX003", "CODX004", "CODX005", "CODX006", "CODX007", "CODX009", "CODX010"} {
		if !hasRule(report, id) {
			t.Fatalf("expected %s in vulnerable fixture findings, got %#v", id, ruleIDs(report))
		}
	}
	diffReport, err := AuditPath(filepath.Join("..", "..", "fixtures", "vulnerable"), AuditOptions{
		All:          true,
		DiffMode:     true,
		ChangedFiles: []string{".github/codex/prompts/review.md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(diffReport, "CODX008") {
		t.Fatalf("expected CODX008 in vulnerable diff fixture findings, got %#v", ruleIDs(diffReport))
	}
}

func TestSecureFixtureHasNoHighOrCriticalFindings(t *testing.T) {
	report, err := AuditPath(filepath.Join("..", "..", "fixtures", "secure"), AuditOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range report.Findings {
		if finding.Severity == SeverityHigh || finding.Severity == SeverityCritical {
			t.Fatalf("secure fixture should not have high/critical findings: %#v", report.Findings)
		}
	}
}

func TestVulnerableFixtureJSONAndSARIFRender(t *testing.T) {
	report, err := AuditPath(filepath.Join("..", "..", "fixtures", "vulnerable"), AuditOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	jsonBytes, err := RenderJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	var jsonShape map[string]any
	if err := json.Unmarshal(jsonBytes, &jsonShape); err != nil {
		t.Fatal(err)
	}
	sarifBytes, err := RenderSARIF(report)
	if err != nil {
		t.Fatal(err)
	}
	var sarifShape struct {
		Version string `json:"version"`
		Runs    []struct {
			Results []struct {
				RuleID    string `json:"ruleId"`
				Level     string `json:"level"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
				Properties struct {
					Severity         Severity   `json:"severity"`
					Confidence       Confidence `json:"confidence"`
					Source           string     `json:"source"`
					PromptBoundary   string     `json:"prompt_boundary"`
					PrivilegeContext string     `json:"privilege_context"`
				} `json:"properties"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(sarifBytes, &sarifShape); err != nil {
		t.Fatal(err)
	}
	if sarifShape.Version != "2.1.0" || len(sarifShape.Runs) != 1 {
		t.Fatalf("unexpected SARIF shape: %#v", sarifShape)
	}
	if len(sarifShape.Runs[0].Results) == 0 {
		t.Fatal("expected vulnerable fixture to render SARIF results")
	}
	var sawCodexPromptFinding bool
	for _, result := range sarifShape.Runs[0].Results {
		if result.RuleID != "CODX001" {
			continue
		}
		sawCodexPromptFinding = true
		if result.Level == "" || result.Properties.Severity == "" || result.Properties.Confidence == "" {
			t.Fatalf("CODX001 SARIF result lost level/severity/confidence: %#v", result)
		}
		if result.Properties.Source == "" || result.Properties.PromptBoundary == "" || result.Properties.PrivilegeContext == "" {
			t.Fatalf("CODX001 SARIF result lost guard properties: %#v", result)
		}
		if len(result.Locations) != 1 || result.Locations[0].PhysicalLocation.ArtifactLocation.URI == "" || result.Locations[0].PhysicalLocation.Region.StartLine == 0 {
			t.Fatalf("CODX001 SARIF result lost location: %#v", result)
		}
	}
	if !sawCodexPromptFinding {
		t.Fatalf("expected CODX001 SARIF result, got %#v", sarifShape.Runs[0].Results)
	}
}
