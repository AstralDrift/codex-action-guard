package cli

import (
	"bytes"
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
