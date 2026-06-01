package githubactions

import (
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type PermissionSummary struct {
	Explicit bool
	WriteAll bool
	ReadAll  bool
	Writes   []string
	Reads    []string
}

func ParsePermissions(node *yaml.Node) PermissionSummary {
	if node == nil {
		return PermissionSummary{}
	}
	summary := PermissionSummary{Explicit: true}
	if node.Kind == yaml.ScalarNode {
		switch strings.ToLower(strings.TrimSpace(node.Value)) {
		case "write-all":
			summary.WriteAll = true
		case "read-all":
			summary.ReadAll = true
		}
		return summary
	}
	for _, pair := range pairs(node) {
		switch strings.ToLower(strings.TrimSpace(scalar(pair.Value))) {
		case "write":
			summary.Writes = append(summary.Writes, pair.Key.Value)
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
