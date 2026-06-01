package profiles

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

//go:embed templates/**
var templateFS embed.FS

const threatModelTemplate = "templates/docs/codex-ci-threat-model.md"

type Profile struct {
	Name        string
	Description string
	Workflow    string
	Prompt      string
	Schema      string
}

type profileTemplate struct {
	Name        string
	Description string
	Workflow    string
	Prompt      string
	Schema      string
}

func Names() []string {
	names := make([]string, 0, len(profileCatalog))
	for name := range profileCatalog {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func Generate(name string, out string, force bool) ([]string, error) {
	if out == "" {
		out = "."
	}
	profile, err := loadProfile(name)
	if err != nil {
		return nil, err
	}
	threatModel, err := readTemplate(threatModelTemplate)
	if err != nil {
		return nil, err
	}

	type generatedFile struct {
		contents            string
		skipIfExistsNoForce bool
	}
	files := map[string]generatedFile{
		filepath.Join(out, ".github", "workflows", "codex-"+name+".yml"):       {contents: profile.Workflow},
		filepath.Join(out, ".github", "codex", "prompts", name+".md"):          {contents: profile.Prompt},
		filepath.Join(out, ".github", "codex", "schemas", name+".schema.json"): {contents: profile.Schema},
		filepath.Join(out, "docs", "codex-ci-threat-model.md"):                 {contents: threatModel, skipIfExistsNoForce: true},
	}
	var written []string
	for path, file := range files {
		if !force {
			if _, err := os.Stat(path); err == nil {
				if file.skipIfExistsNoForce {
					continue
				}
				return nil, fmt.Errorf("%s already exists; pass --force to overwrite generated profile files", path)
			}
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(file.contents), 0o644); err != nil {
			return nil, err
		}
		written = append(written, path)
	}
	sort.Strings(written)
	return written, nil
}

func loadProfile(name string) (Profile, error) {
	tmpl, ok := profileCatalog[name]
	if !ok {
		return Profile{}, fmt.Errorf("unknown profile %q", name)
	}
	workflow, err := readTemplate(tmpl.Workflow)
	if err != nil {
		return Profile{}, err
	}
	prompt, err := readTemplate(tmpl.Prompt)
	if err != nil {
		return Profile{}, err
	}
	schema, err := readTemplate(tmpl.Schema)
	if err != nil {
		return Profile{}, err
	}
	return Profile{
		Name:        tmpl.Name,
		Description: tmpl.Description,
		Workflow:    workflow,
		Prompt:      prompt,
		Schema:      schema,
	}, nil
}

func readTemplate(path string) (string, error) {
	data, err := templateFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read embedded profile template %s: %w", path, err)
	}
	return string(data), nil
}

var profileCatalog = map[string]profileTemplate{
	"pr-review-readonly": {
		Name:        "pr-review-readonly",
		Description: "Read-only PR review profile that produces a JSON artifact.",
		Workflow:    "templates/workflows/pr-review-readonly.yml",
		Prompt:      "templates/prompts/pr-review-readonly.md",
		Schema:      "templates/schemas/review.schema.json",
	},
	"ci-failure-analysis-readonly": {
		Name:        "ci-failure-analysis-readonly",
		Description: "Read-only CI failure analysis profile for workflow_run events.",
		Workflow:    "templates/workflows/ci-failure-analysis-readonly.yml",
		Prompt:      "templates/prompts/ci-failure-analysis-readonly.md",
		Schema:      "templates/schemas/review.schema.json",
	},
	"release-notes-draft": {
		Name:        "release-notes-draft",
		Description: "Manual read-only release notes drafting profile.",
		Workflow:    "templates/workflows/release-notes-draft.yml",
		Prompt:      "templates/prompts/release-notes-draft.md",
		Schema:      "templates/schemas/release.schema.json",
	},
	"security-review-readonly": {
		Name:        "security-review-readonly",
		Description: "Read-only security review profile for PRs or manual runs.",
		Workflow:    "templates/workflows/security-review-readonly.yml",
		Prompt:      "templates/prompts/security-review-readonly.md",
		Schema:      "templates/schemas/review.schema.json",
	},
	"label-gated-maintainer-task": {
		Name:        "label-gated-maintainer-task",
		Description: "Maintainer-label gated profile for privileged follow-up review packets.",
		Workflow:    "templates/workflows/label-gated-maintainer-task.yml",
		Prompt:      "templates/prompts/label-gated-maintainer-task.md",
		Schema:      "templates/schemas/review.schema.json",
	},
}
