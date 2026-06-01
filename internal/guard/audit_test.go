package guard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AstralDrift/codex-action-guard/internal/profiles"
)

func TestAuditDetectsUnsafePromptAndSinks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".github/workflows/unsafe.yml", `name: unsafe
on:
  pull_request_target:
    types: [opened]
permissions: write-all
jobs:
  codex:
    runs-on: ubuntu-latest
    env:
      OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
    steps:
      - uses: actions/checkout@v5
        with:
          ref: ${{ github.event.pull_request.head.sha }}
      - name: Run Codex
        id: run_codex
        uses: openai/codex-action@v1
        with:
          openai-api-key: ${{ secrets.OPENAI_API_KEY }}
          prompt: |
            Review this pull request:
            ${{ github.event.pull_request.body }}
          sandbox: danger-full-access
      - name: Post comment
        uses: actions/github-script@v7
        env:
          CODEX_FINAL_MESSAGE: ${{ steps.run_codex.outputs.final-message }}
        with:
          script: |
            await github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.payload.pull_request.number,
              body: process.env.CODEX_FINAL_MESSAGE,
            })
`)

	report, err := AuditPath(root, AuditOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"CODX001", "CODX002", "CODX003", "CODX004", "CODX005", "CODX006", "CODX007", "CODX009", "CODX010"} {
		if !hasRule(report, id) {
			t.Fatalf("expected %s in findings, got %#v", id, ruleIDs(report))
		}
	}
}

func TestGeneratedReadOnlyProfileAuditsCleanly(t *testing.T) {
	root := t.TempDir()
	if _, err := profiles.Generate("pr-review-readonly", root, false); err != nil {
		t.Fatal(err)
	}
	report, err := AuditPath(root, AuditOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if report.MeetsThreshold(SeverityHigh) {
		t.Fatalf("generated read-only profile should not produce high findings: %#v", report.Findings)
	}
	if len(report.CodexInvocations) != 1 {
		t.Fatalf("expected one Codex invocation, got %d", len(report.CodexInvocations))
	}
	if len(report.SafePatterns) == 0 {
		t.Fatal("expected safe patterns for generated profile")
	}
}

func TestDiffModeDetectsChangedPromptReference(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".github/workflows/codex.yml", `name: codex
on:
  pull_request:
permissions:
  contents: read
jobs:
  codex:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@v5
        with:
          persist-credentials: false
      - id: run_codex
        uses: openai/codex-action@v1
        with:
          openai-api-key: ${{ secrets.OPENAI_API_KEY }}
          prompt-file: .github/codex/prompts/review.md
          output-schema-file: .github/codex/schemas/review.schema.json
          sandbox: read-only
`)
	writeFile(t, root, ".github/codex/prompts/review.md", "trusted prompt\n")
	writeFile(t, root, ".github/codex/schemas/review.schema.json", "{}\n")

	report, err := AuditPath(root, AuditOptions{
		All:          true,
		DiffMode:     true,
		ChangedFiles: []string{".github/codex/prompts/review.md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(report, "CODX008") {
		t.Fatalf("expected CODX008 in findings, got %#v", ruleIDs(report))
	}
}

func TestAuditDetectsDirectCodexExec(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".github/workflows/direct.yml", `name: direct
on:
  issue_comment:
permissions:
  contents: read
jobs:
  codex:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - name: Direct Codex
        run: |
          codex exec --sandbox danger-full-access "${{ github.event.comment.body }}"
`)

	report, err := AuditPath(root, AuditOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(report, "CODX001") {
		t.Fatalf("expected CODX001 for direct codex exec, got %#v", ruleIDs(report))
	}
	if !hasRule(report, "CODX004") {
		t.Fatalf("expected CODX004 for direct codex exec, got %#v", ruleIDs(report))
	}
}

func writeFile(t *testing.T, root string, rel string, contents string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasRule(report Report, id string) bool {
	for _, finding := range report.Findings {
		if finding.RuleID == id {
			return true
		}
	}
	return false
}

func ruleIDs(report Report) []string {
	var ids []string
	for _, finding := range report.Findings {
		ids = append(ids, finding.RuleID)
	}
	return ids
}
