package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

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
