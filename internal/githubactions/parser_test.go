package githubactions

import "testing"

func TestParseWorkflow(t *testing.T) {
	workflow, err := ParseWorkflow("workflow.yml", []byte(`name: test
on:
  pull_request:
  workflow_dispatch:
permissions:
  contents: read
jobs:
  codex:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - name: Run Codex
        id: run_codex
        uses: openai/codex-action@v1
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(workflow.Triggers) != 2 {
		t.Fatalf("expected 2 triggers, got %#v", workflow.Triggers)
	}
	if len(workflow.Jobs) != 1 || workflow.Jobs[0].ID != "codex" {
		t.Fatalf("unexpected jobs: %#v", workflow.Jobs)
	}
	if !workflow.Jobs[0].Permissions.Explicit {
		t.Fatal("expected explicit job permissions")
	}
}

func TestParseWorkflowAllowsJobsWithoutSteps(t *testing.T) {
	workflow, err := ParseWorkflow("workflow.yml", []byte(`name: no steps
on: workflow_dispatch
jobs:
  metadata:
    runs-on: ubuntu-latest
    permissions:
      contents: read
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(workflow.Jobs) != 1 {
		t.Fatalf("expected one job, got %#v", workflow.Jobs)
	}
	if len(workflow.Jobs[0].Steps) != 0 {
		t.Fatalf("expected no steps, got %#v", workflow.Jobs[0].Steps)
	}
	if !workflow.Jobs[0].Permissions.Explicit {
		t.Fatal("expected explicit permissions to still be parsed")
	}
}

func TestContainsUntrustedExpression(t *testing.T) {
	if !ContainsUntrustedExpression("${{ github.event.pull_request.body }}") {
		t.Fatal("expected PR body to be untrusted")
	}
	if ContainsUntrustedExpression("${{ github.event.pull_request.number }}") {
		t.Fatal("PR number should not be treated as untrusted text")
	}
}

func TestRelevantDiffFiles(t *testing.T) {
	files := RelevantDiffFiles([]string{"README.md", ".github/workflows/codex.yml", ".github/codex/prompts/review.md"})
	if len(files) != 2 {
		t.Fatalf("expected two relevant files, got %#v", files)
	}
}
