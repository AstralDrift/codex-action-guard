package installer

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed templates/*.yml
var templateFS embed.FS

const (
	PresetArtifact = "artifact"
	PresetSARIF    = "sarif"
	WorkflowPath   = ".github/workflows/codex-action-guard.yml"
)

var ErrUnknownPreset = errors.New("unknown install preset")

type Options struct {
	Preset string
	Out    string
	Force  bool
}

type Result struct {
	Path   string
	Preset string
}

func Presets() []string {
	presets := make([]string, 0, len(presetTemplates))
	for preset := range presetTemplates {
		presets = append(presets, preset)
	}
	sort.Strings(presets)
	return presets
}

func IsUnknownPreset(err error) bool {
	return errors.Is(err, ErrUnknownPreset)
}

func Generate(opts Options) (Result, error) {
	preset := normalizePreset(opts.Preset)
	if opts.Out == "" {
		opts.Out = "."
	}
	template, err := Template(preset)
	if err != nil {
		return Result{}, err
	}
	path := filepath.Join(opts.Out, filepath.FromSlash(WorkflowPath))
	if !opts.Force {
		if _, err := os.Stat(path); err == nil {
			return Result{}, fmt.Errorf("%s already exists; pass --force to overwrite generated install workflow", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return Result{}, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(path, []byte(template), 0o644); err != nil {
		return Result{}, err
	}
	return Result{Path: path, Preset: preset}, nil
}

func Template(preset string) (string, error) {
	preset = normalizePreset(preset)
	path, ok := presetTemplates[preset]
	if !ok {
		return "", fmt.Errorf("%w %q", ErrUnknownPreset, preset)
	}
	data, err := templateFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read embedded install template %s: %w", path, err)
	}
	return string(data), nil
}

func normalizePreset(preset string) string {
	preset = strings.ToLower(strings.TrimSpace(preset))
	if preset == "" {
		return PresetArtifact
	}
	return preset
}

var presetTemplates = map[string]string{
	PresetArtifact: "templates/artifact.yml",
	PresetSARIF:    "templates/sarif.yml",
}
