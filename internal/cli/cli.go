package cli

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/AstralDrift/codex-action-guard/internal/githubactions"
	"github.com/AstralDrift/codex-action-guard/internal/guard"
	"github.com/AstralDrift/codex-action-guard/internal/profiles"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

func Run(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	switch args[0] {
	case "version":
		return runVersion(stdout, build)
	case "init":
		return runInit(args[1:], stdout, stderr, build)
	case "audit":
		return runAudit(args[1:], stdout, stderr, build)
	case "diff":
		return runDiff(args[1:], stdout, stderr, build)
	case "packet":
		return runPacket(args[1:], stdout, stderr, build)
	case "rules":
		return runRules(args[1:], stdout, stderr, build)
	case "explain":
		return runExplain(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runVersion(stdout io.Writer, build BuildInfo) int {
	if build.Version == "" {
		build.Version = "dev"
	}
	if build.Commit == "" {
		build.Commit = "unknown"
	}
	if build.Date == "" {
		build.Date = "unknown"
	}
	fmt.Fprintf(stdout, "%s %s\n", guard.ToolName, build.Version)
	fmt.Fprintf(stdout, "commit: %s\n", build.Commit)
	fmt.Fprintf(stdout, "built: %s\n", build.Date)
	fmt.Fprintf(stdout, "go: %s\n", runtime.Version())
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Fprintf(stdout, "module: %s\n", info.Main.Path)
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" || setting.Key == "vcs.time" || setting.Key == "vcs.modified" {
				fmt.Fprintf(stdout, "%s: %s\n", setting.Key, setting.Value)
			}
		}
	}
	return 0
}

func runInit(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	profile := fs.String("profile", "", "profile name")
	out := fs.String("out", ".", "output repository path")
	force := fs.Bool("force", false, "overwrite generated files")
	if err := fs.Parse(normalizeFlagArgs(args, map[string]bool{
		"profile": true,
		"out":     true,
	})); err != nil {
		return 2
	}
	if *profile == "" {
		fmt.Fprintf(stderr, "init requires --profile. Available profiles: %s\n", strings.Join(profiles.Names(), ", "))
		return 2
	}
	written, err := profiles.Generate(*profile, *out, *force)
	if err != nil {
		fmt.Fprintf(stderr, "init failed: %v\n", err)
		return 1
	}
	for _, path := range written {
		fmt.Fprintf(stdout, "created %s\n", path)
	}
	report, err := guard.AuditPath(*out, guard.AuditOptions{All: true, ToolVersion: build.Version})
	if err != nil {
		fmt.Fprintf(stderr, "generated output audit failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "validation: scanned %d files, found %d findings\n", len(report.ScannedFiles), len(report.Findings))
	if report.MeetsThreshold(guard.SeverityCritical) {
		fmt.Fprintln(stderr, "generated profile produced critical findings")
		return 1
	}
	return 0
}

func runAudit(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	fs := flag.NewFlagSet("audit", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "markdown", "markdown, json, or sarif")
	output := fs.String("output", "", "write report to file")
	failOn := fs.String("fail-on", "none", "info, low, medium, high, critical, or none")
	all := fs.Bool("all", false, "include Codex prompt/schema/AGENTS files in scanned file list")
	if err := fs.Parse(normalizeFlagArgs(args, map[string]bool{
		"format":  true,
		"output":  true,
		"fail-on": true,
	})); err != nil {
		return 2
	}
	target := "."
	if fs.NArg() > 0 {
		target = fs.Arg(0)
	}
	threshold, ok := guard.ParseSeverity(*failOn)
	if !ok {
		fmt.Fprintf(stderr, "invalid --fail-on value %q\n", *failOn)
		return 2
	}
	report, err := guard.AuditPath(target, guard.AuditOptions{All: *all, ToolVersion: build.Version})
	if err != nil {
		fmt.Fprintf(stderr, "audit failed: %v\n", err)
		return 1
	}
	if err := emitReport(report, *format, *output, stdout); err != nil {
		fmt.Fprintf(stderr, "audit output failed: %v\n", err)
		return 1
	}
	if report.MeetsThreshold(threshold) {
		return 3
	}
	return 0
}

func runDiff(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "markdown", "markdown, json, or sarif")
	output := fs.String("output", "", "write report to file")
	failOn := fs.String("fail-on", "none", "info, low, medium, high, critical, or none")
	if err := fs.Parse(normalizeFlagArgs(args, map[string]bool{
		"format":  true,
		"output":  true,
		"fail-on": true,
	})); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "diff requires a git rev range, for example main...HEAD")
		return 2
	}
	threshold, ok := guard.ParseSeverity(*failOn)
	if !ok {
		fmt.Fprintf(stderr, "invalid --fail-on value %q\n", *failOn)
		return 2
	}
	changed, err := gitChangedFiles(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "diff failed: %v\n", err)
		return 1
	}
	relevant := filterRelevant(changed)
	if len(relevant) == 0 {
		report := guard.NewReport(currentDir(), build.Version)
		report.ScannedFiles = []string{}
		report.ProfileSuggestions = []string{"No Codex-relevant workflow, prompt, schema, or AGENTS.md files changed."}
		if err := emitReport(report, *format, *output, stdout); err != nil {
			fmt.Fprintf(stderr, "diff output failed: %v\n", err)
			return 1
		}
		return 0
	}
	report, err := guard.AuditPath(".", guard.AuditOptions{All: true, DiffMode: true, ChangedFiles: relevant, ToolVersion: build.Version})
	if err != nil {
		fmt.Fprintf(stderr, "diff audit failed: %v\n", err)
		return 1
	}
	if err := emitReport(report, *format, *output, stdout); err != nil {
		fmt.Fprintf(stderr, "diff output failed: %v\n", err)
		return 1
	}
	if report.MeetsThreshold(threshold) {
		return 3
	}
	return 0
}

func runPacket(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	fs := flag.NewFlagSet("packet", flag.ContinueOnError)
	fs.SetOutput(stderr)
	target := fs.String("target", "human", "human or codex")
	changedRange := fs.String("changed", "", "git rev range to summarize")
	output := fs.String("output", "", "write packet to file")
	if err := fs.Parse(normalizeFlagArgs(args, map[string]bool{
		"target":  true,
		"changed": true,
		"output":  true,
	})); err != nil {
		return 2
	}
	if *target != "human" && *target != "codex" {
		fmt.Fprintln(stderr, "--target must be human or codex")
		return 2
	}
	var changed []string
	var report guard.Report
	var err error
	if *changedRange != "" {
		changed, err = gitChangedFiles(*changedRange)
		if err != nil {
			fmt.Fprintf(stderr, "packet failed: %v\n", err)
			return 1
		}
		report, err = guard.AuditPath(".", guard.AuditOptions{All: true, DiffMode: true, ChangedFiles: filterRelevant(changed), ToolVersion: build.Version})
	} else {
		report, err = guard.AuditPath(".", guard.AuditOptions{All: true, ToolVersion: build.Version})
	}
	if err != nil {
		fmt.Fprintf(stderr, "packet audit failed: %v\n", err)
		return 1
	}
	packet := guard.RenderPacket(report, *target, changed)
	if *output != "" {
		if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil && filepath.Dir(*output) != "." {
			fmt.Fprintf(stderr, "packet output failed: %v\n", err)
			return 1
		}
		if err := os.WriteFile(*output, []byte(packet), 0o644); err != nil {
			fmt.Fprintf(stderr, "packet output failed: %v\n", err)
			return 1
		}
		return 0
	}
	fmt.Fprint(stdout, packet)
	return 0
}

func runExplain(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "explain requires a rule id, for example CODX001")
		return 2
	}
	text, err := guard.ExplainRule(strings.ToUpper(args[0]))
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	fmt.Fprint(stdout, text)
	return 0
}

func runRules(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	fs := flag.NewFlagSet("rules", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "json", "json or markdown")
	output := fs.String("output", "", "write rule catalog to file")
	if err := fs.Parse(normalizeFlagArgs(args, map[string]bool{
		"format": true,
		"output": true,
	})); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "rules does not accept positional arguments")
		return 2
	}

	var data []byte
	switch strings.ToLower(*format) {
	case "json":
		var err error
		data, err = guard.RenderRulesJSON(build.Version)
		if err != nil {
			fmt.Fprintf(stderr, "rules output failed: %v\n", err)
			return 1
		}
		data = append(data, '\n')
	case "markdown", "md":
		data = []byte(guard.RenderRulesMarkdown(build.Version))
	default:
		fmt.Fprintf(stderr, "unknown rules format %q\n", *format)
		return 2
	}
	if err := writeOutput(data, *output, stdout); err != nil {
		fmt.Fprintf(stderr, "rules output failed: %v\n", err)
		return 1
	}
	return 0
}

func emitReport(report guard.Report, format string, output string, stdout io.Writer) error {
	var data []byte
	switch strings.ToLower(format) {
	case "markdown", "md":
		data = []byte(guard.RenderMarkdown(report))
	case "json":
		var err error
		data, err = guard.RenderJSON(report)
		if err != nil {
			return err
		}
		data = append(data, '\n')
	case "sarif":
		var err error
		data, err = guard.RenderSARIF(report)
		if err != nil {
			return err
		}
		data = append(data, '\n')
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	return writeOutput(data, output, stdout)
}

func writeOutput(data []byte, output string, stdout io.Writer) error {
	if output == "" {
		_, err := stdout.Write(data)
		return err
	}
	if dir := filepath.Dir(output); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(output, data, 0o644)
}

func gitChangedFiles(revRange string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", revRange, "--")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("could not read git diff for %q: %s", revRange, msg)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, filepath.ToSlash(line))
		}
	}
	return files, nil
}

func filterRelevant(files []string) []string {
	return githubactions.RelevantDiffFiles(files)
}

func currentDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `codex-action-guard - safe-by-default Codex GitHub Action workflows

Usage:
  codex-action-guard version
  codex-action-guard init --profile <name> [--out <repo>] [--force]
  codex-action-guard audit [path] [--format markdown|json|sarif] [--output <file>] [--fail-on low|medium|high|critical|none] [--all]
  codex-action-guard diff <rev-range> [--format markdown|json|sarif] [--output <file>] [--fail-on low|medium|high|critical|none]
  codex-action-guard packet [--target codex|human] [--changed <rev-range>] [--output <file>]
  codex-action-guard rules [--format json|markdown] [--output <file>]
  codex-action-guard explain <RULE_ID>

Profiles:
  %s
`, strings.Join(profiles.Names(), ", "))
}
