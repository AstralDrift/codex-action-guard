package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateAllProfiles(t *testing.T) {
	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			written, err := Generate(name, root, false)
			if err != nil {
				t.Fatal(err)
			}
			if len(written) != 4 {
				t.Fatalf("expected workflow, prompt, schema, and threat model, got %#v", written)
			}
			for _, rel := range []string{
				filepath.Join(".github", "workflows", "codex-"+name+".yml"),
				filepath.Join(".github", "codex", "prompts", name+".md"),
				filepath.Join(".github", "codex", "schemas", name+".schema.json"),
				filepath.Join("docs", "codex-ci-threat-model.md"),
			} {
				if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
					t.Fatalf("expected generated file %s: %v", rel, err)
				}
			}
		})
	}
}

func TestGenerateSkipsExistingThreatModelForAdditionalProfiles(t *testing.T) {
	root := t.TempDir()
	if _, err := Generate("pr-review-readonly", root, false); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate("security-review-readonly", root, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "workflows", "codex-security-review-readonly.yml")); err != nil {
		t.Fatal(err)
	}
}
