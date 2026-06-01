package githubactions

import (
	"path/filepath"
	"strings"
)

func IsCodexRelevantPath(path string) bool {
	path = filepath.ToSlash(path)
	if IsWorkflowPath(path) {
		return true
	}
	return IsPromptOrSchemaPath(path) ||
		path == "AGENTS.md"
}

func IsWorkflowPath(path string) bool {
	path = filepath.ToSlash(path)
	return strings.HasPrefix(path, ".github/workflows/") && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml"))
}

func IsPromptOrSchemaPath(path string) bool {
	path = filepath.ToSlash(path)
	return strings.HasPrefix(path, ".github/codex/prompts/") || strings.HasPrefix(path, ".github/codex/schemas/")
}

func NormalizeWorkflowRef(ref string) string {
	ref = strings.TrimSpace(strings.Trim(ref, `"'`))
	ref = strings.TrimPrefix(ref, "./")
	return filepath.ToSlash(ref)
}

func RelevantDiffFiles(files []string) []string {
	var out []string
	for _, file := range files {
		if IsCodexRelevantPath(file) {
			out = append(out, filepath.ToSlash(file))
		}
	}
	return out
}
