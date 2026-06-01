package githubactions

import (
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type PermissionSummary struct {
	Explicit bool
	Line     int
	Missing  bool
	WriteAll bool
	ReadAll  bool
	Writes   []string
	Reads    []string
	IDToken  bool
}

func ParsePermissions(node *yaml.Node) PermissionSummary {
	if node == nil {
		return PermissionSummary{Missing: true}
	}
	summary := PermissionSummary{Explicit: true, Line: node.Line}
	if node.Kind == yaml.ScalarNode {
		switch strings.ToLower(strings.TrimSpace(node.Value)) {
		case "write-all":
			summary.WriteAll = true
		case "read-all":
			summary.ReadAll = true
		case "{}":
			summary.Reads = append(summary.Reads, "none")
		}
		return summary
	}
	for _, pair := range Pairs(node) {
		switch strings.ToLower(strings.TrimSpace(Scalar(pair.Value))) {
		case "write":
			summary.Writes = append(summary.Writes, pair.Key.Value)
			if pair.Key.Value == "id-token" {
				summary.IDToken = true
			}
		case "read":
			summary.Reads = append(summary.Reads, pair.Key.Value)
		}
	}
	sort.Strings(summary.Writes)
	sort.Strings(summary.Reads)
	return summary
}

func (p PermissionSummary) WriteCapable() bool {
	return p.WriteAll || len(p.Writes) > 0
}

func (p PermissionSummary) HasBroadWrites() bool {
	if p.WriteAll {
		return true
	}
	broad := map[string]bool{
		"actions":         true,
		"attestations":    true,
		"checks":          true,
		"contents":        true,
		"deployments":     true,
		"id-token":        true,
		"packages":        true,
		"pages":           true,
		"security-events": true,
	}
	for _, scope := range p.Writes {
		if broad[scope] {
			return true
		}
	}
	return false
}

func (p PermissionSummary) Description() string {
	switch {
	case p.Missing:
		return "missing explicit permissions; repository GITHUB_TOKEN defaults may apply"
	case p.WriteAll:
		return "permissions: write-all"
	case len(p.Writes) > 0:
		return "write permissions: " + strings.Join(p.Writes, ", ")
	case p.ReadAll:
		return "permissions: read-all"
	case len(p.Reads) > 0:
		return "read permissions: " + strings.Join(p.Reads, ", ")
	default:
		return "permissions explicitly empty or none"
	}
}
