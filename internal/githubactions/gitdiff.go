package githubactions

import (
	"path/filepath"
	"strings"
)

func IsCodexRelevantPath(path string) bool {
	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, ".github/workflows/") && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
		return true
	}
	return strings.HasPrefix(path, ".github/codex/prompts/") ||
		strings.HasPrefix(path, ".github/codex/schemas/") ||
		path == "AGENTS.md"
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
