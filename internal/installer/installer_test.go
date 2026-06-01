package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AstralDrift/codex-action-guard/internal/githubactions"
	"github.com/AstralDrift/codex-action-guard/internal/guard"
)

func TestGeneratedWorkflowsParseAndAuditCleanly(t *testing.T) {
	for _, preset := range Presets() {
		t.Run(preset, func(t *testing.T) {
			root := t.TempDir()
			result, err := Generate(Options{Preset: preset, Out: root})
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(result.Path)
			if err != nil {
				t.Fatal(err)
			}
			workflow, err := githubactions.ParseWorkflow(filepath.ToSlash(WorkflowPath), data)
			if err != nil {
				t.Fatal(err)
			}
			if len(workflow.Jobs) != 1 {
				t.Fatalf("expected one job, got %d", len(workflow.Jobs))
			}
			report, err := guard.AuditPath(root, guard.AuditOptions{All: true, ToolVersion: "test"})
			if err != nil {
				t.Fatal(err)
			}
			if report.MeetsThreshold(guard.SeverityHigh) {
				t.Fatalf("expected no high findings for %s preset: %#v", preset, report.Findings)
			}
		})
	}
}

func TestArtifactTemplateSafetyProperties(t *testing.T) {
	template, err := Template(PresetArtifact)
	if err != nil {
		t.Fatal(err)
	}
	mustContainAll(t, template,
		"uses: actions/checkout@v5",
		"persist-credentials: false",
		"uses: AstralDrift/codex-action-guard@v0",
		"fail-on: high",
		"format: markdown",
		"uses: actions/upload-artifact@v5",
		"FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true",
		"permissions:\n  contents: read",
		"workflow_dispatch:",
		`- ".github/workflows/**"`,
		`- ".github/codex/**"`,
		`- "AGENTS.md"`,
		`- "action.yml"`,
	)
	if strings.Contains(template, "security-events: write") {
		t.Fatalf("artifact preset should not request security-events write permission:\n%s", template)
	}
}

func TestSARIFTemplateSafetyProperties(t *testing.T) {
	template, err := Template(PresetSARIF)
	if err != nil {
		t.Fatal(err)
	}
	mustContainAll(t, template,
		"uses: actions/checkout@v5",
		"persist-credentials: false",
		"uses: AstralDrift/codex-action-guard@v0",
		"fail-on: high",
		"format: sarif",
		"security-events: write",
		"uses: github/codeql-action/upload-sarif@v4",
		"FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true",
		"sarif_file: codex-action-guard.sarif",
		"workflow_dispatch:",
	)
}

func TestExamplesMatchEmbeddedTemplates(t *testing.T) {
	tests := []struct {
		preset string
		path   string
	}{
		{preset: PresetArtifact, path: "../../examples/install/codex-action-guard-artifact.yml"},
		{preset: PresetSARIF, path: "../../examples/install/codex-action-guard-sarif.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			template, err := Template(tt.preset)
			if err != nil {
				t.Fatal(err)
			}
			example, err := os.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if string(example) != template {
				t.Fatalf("example %s does not match embedded template for %s", tt.path, tt.preset)
			}
		})
	}
}

func mustContainAll(t *testing.T, text string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected template to contain %q:\n%s", needle, text)
		}
	}
}
