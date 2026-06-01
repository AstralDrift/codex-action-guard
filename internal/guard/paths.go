package guard

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AstralDrift/codex-action-guard/internal/githubactions"
)

func findRepoRoot(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			return start
		}
		dir = next
	}
}

func inferWorkflowRoot(absFile string) (string, bool) {
	parts := strings.Split(filepath.ToSlash(absFile), "/")
	for i := len(parts) - 3; i >= 0; i-- {
		if parts[i] == ".github" && i+1 < len(parts) && parts[i+1] == "workflows" {
			root := filepath.FromSlash(strings.Join(parts[:i], "/"))
			if root == "" {
				root = string(filepath.Separator)
			}
			return root, true
		}
	}
	return "", false
}

func isWorkflowFile(rel string) bool {
	return githubactions.IsWorkflowPath(rel)
}

func isRelevantFile(rel string) bool {
	return githubactions.IsCodexRelevantPath(rel)
}

func isPromptOrSchemaFile(rel string) bool {
	return githubactions.IsPromptOrSchemaPath(rel)
}
