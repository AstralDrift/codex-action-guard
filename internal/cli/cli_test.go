package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAuditAcceptsFlagsAfterPath(t *testing.T) {
	root := t.TempDir()
	workflow := filepath.Join(root, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(workflow), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(workflow, []byte("name: ci\non: push\njobs: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"audit", root, "--all", "--format", "json"}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"tool": "codex-action-guard"`) {
		t.Fatalf("expected JSON report, got %s", stdout.String())
	}
}

func TestNormalizeFlagArgsSupportsInterspersedFlagsAndSeparator(t *testing.T) {
	got := normalizeFlagArgs(
		[]string{"target", "--format", "json", "--all", "--", "--literal"},
		map[string]bool{"format": true},
	)
	want := []string{"--format", "json", "--all", "target", "--", "--literal"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalized args\nwant %#v\ngot  %#v", want, got)
	}
}

func TestRulesJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"rules", "--format", "json"}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	var catalog struct {
		Metadata struct {
			Tool        string `json:"tool"`
			Version     string `json:"version"`
			RuleVersion string `json:"rule_version"`
		} `json:"metadata"`
		Rules []struct {
			ID              string `json:"id"`
			DefaultSeverity string `json:"default_severity"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	if catalog.Metadata.Tool != "codex-action-guard" || catalog.Metadata.Version != "test" || catalog.Metadata.RuleVersion == "" {
		t.Fatalf("unexpected metadata: %#v", catalog.Metadata)
	}
	if len(catalog.Rules) != 10 {
		t.Fatalf("expected 10 rules, got %#v", catalog.Rules)
	}
	wantIDs := []string{"CODX001", "CODX002", "CODX003", "CODX004", "CODX005", "CODX006", "CODX007", "CODX008", "CODX009", "CODX010"}
	for i, rule := range catalog.Rules {
		want := wantIDs[i]
		if rule.ID != want || rule.DefaultSeverity == "" {
			t.Fatalf("unexpected rule at %d: %#v", i, rule)
		}
	}
	if strings.Contains(stdout.String(), "generated_at") {
		t.Fatalf("rules export should be deterministic and omit timestamps: %s", stdout.String())
	}
}

func TestRulesMarkdown(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"rules", "--format", "markdown"}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	for _, id := range []string{"CODX001", "CODX002", "CODX003", "CODX004", "CODX005", "CODX006", "CODX007", "CODX008", "CODX009", "CODX010"} {
		if !strings.Contains(stdout.String(), id) {
			t.Fatalf("expected markdown rules output to contain %s: %s", id, stdout.String())
		}
	}
	if !strings.Contains(stdout.String(), "| `CODX001` | medium |") {
		t.Fatalf("expected markdown rules output to include default severities: %s", stdout.String())
	}
}

func TestRulesOutputFile(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "rules.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"rules", "--output", out}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout when --output is set, got %s", stdout.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"rules"`) {
		t.Fatalf("expected rules JSON in output file, got %s", string(data))
	}
}

func TestRulesInvalidFormat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"rules", "--format", "xml"}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown rules format "xml"`) {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestInstallCreatesDefaultArtifactWorkflow(t *testing.T) {
	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"install", "--out", root}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "preset: artifact") {
		t.Fatalf("expected selected preset in stdout, got %s", stdout.String())
	}
	data := readInstalledWorkflow(t, root)
	text := string(data)
	for _, want := range []string{
		"uses: AstralDrift/codex-action-guard@v0",
		"format: markdown",
		"uses: actions/upload-artifact@v5",
		"persist-credentials: false",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected workflow to contain %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "security-events: write") {
		t.Fatalf("artifact preset should not request security-events write permission:\n%s", text)
	}
}

func TestInstallCreatesSARIFWorkflow(t *testing.T) {
	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"install", "--preset", "sarif", "--out", root}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "preset: sarif") {
		t.Fatalf("expected selected preset in stdout, got %s", stdout.String())
	}
	text := string(readInstalledWorkflow(t, root))
	for _, want := range []string{
		"uses: AstralDrift/codex-action-guard@v0",
		"format: sarif",
		"security-events: write",
		"uses: github/codeql-action/upload-sarif@v4",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected workflow to contain %q:\n%s", want, text)
		}
	}
}

func TestInstallOutWritesUnderTargetRepo(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "repo")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"install", "--out", target}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	want := filepath.Join(target, ".github", "workflows", "codex-action-guard.yml")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected workflow under --out path: %v", err)
	}
	if !strings.Contains(stdout.String(), want) {
		t.Fatalf("expected stdout to include created file path %q, got %s", want, stdout.String())
	}
}

func TestInstallRefusesOverwriteWithoutForce(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".github", "workflows", "codex-action-guard.yml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"install", "--out", root}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing\n" {
		t.Fatalf("expected existing workflow to be preserved, got %s", string(data))
	}
	if !strings.Contains(stderr.String(), "pass --force") {
		t.Fatalf("expected force hint, got %s", stderr.String())
	}
}

func TestInstallForceOverwrites(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".github", "workflows", "codex-action-guard.yml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"install", "--out", root, "--force"}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "existing\n" || !strings.Contains(string(data), "Codex Action Guard") {
		t.Fatalf("expected generated workflow to overwrite existing content, got %s", string(data))
	}
}

func TestInstallUnknownPreset(t *testing.T) {
	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"install", "--preset", "junit", "--out", root}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown install preset "junit"`) {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func readInstalledWorkflow(t *testing.T, root string) []byte {
	t.Helper()
	path := filepath.Join(root, ".github", "workflows", "codex-action-guard.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCLISmokeAgainstFixtures(t *testing.T) {
	fixture := filepath.Join("..", "..", "fixtures", "vulnerable")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"audit", fixture, "--all", "--format", "json", "--fail-on", "critical"}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 3 {
		t.Fatalf("expected threshold exit 3, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"rule_id": "CODX001"`) {
		t.Fatalf("expected fixture findings in JSON report, got %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"explain", "CODX001"}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("explain failed with %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Untrusted GitHub event content") {
		t.Fatalf("unexpected explain output: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join("..", "..", "fixtures", "vulnerable")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
	code = Run([]string{"packet", "--target", "codex"}, &stdout, &stderr, BuildInfo{Version: "test"})
	if code != 0 {
		t.Fatalf("packet failed with %d: %s", code, stderr.String())
	}
	if !strings.Contains(strings.ToLower(stdout.String()), "do not invent vulnerabilities") {
		t.Fatalf("expected codex packet instructions, got %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "CODX001") {
		t.Fatalf("expected fixture findings in packet, got %s", stdout.String())
	}
}
