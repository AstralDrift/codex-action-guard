package codex

import "github.com/AstralDrift/codex-action-guard/internal/profiles"

func ProfileNames() []string {
	return profiles.Names()
}

func GenerateProfile(name string, out string, force bool) ([]string, error) {
	return profiles.Generate(name, out, force)
}
