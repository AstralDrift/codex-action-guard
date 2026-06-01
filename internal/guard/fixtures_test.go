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
		Runs    []any  `json:"runs"`
	}
	if err := json.Unmarshal(sarifBytes, &sarifShape); err != nil {
		t.Fatal(err)
	}
	if sarifShape.Version != "2.1.0" || len(sarifShape.Runs) != 1 {
		t.Fatalf("unexpected SARIF shape: %#v", sarifShape)
	}
}
