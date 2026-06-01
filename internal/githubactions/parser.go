package githubactions

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

func ParseWorkflow(file string, data []byte) (Workflow, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return Workflow{}, fmt.Errorf("%s: %w", file, err)
	}
	if len(doc.Content) == 0 {
		return Workflow{}, fmt.Errorf("%s: empty workflow", file)
	}
	root := doc.Content[0]
	workflow := Workflow{File: file, Triggers: parseTriggers(lookup(root, "on"))}
	jobs := lookup(root, "jobs")
	for _, pair := range pairs(jobs) {
		job := Job{ID: pair.Key.Value, Line: pair.Value.Line, Permissions: ParsePermissions(lookup(pair.Value, "permissions"))}
		steps := lookup(pair.Value, "steps")
		for _, stepNode := range steps.Content {
			if stepNode.Kind != yaml.MappingNode {
				continue
			}
			job.Steps = append(job.Steps, Step{
				Name: scalar(lookup(stepNode, "name")),
				ID:   scalar(lookup(stepNode, "id")),
				Uses: scalar(lookup(stepNode, "uses")),
				Run:  scalar(lookup(stepNode, "run")),
				Line: stepNode.Line,
			})
		}
		workflow.Jobs = append(workflow.Jobs, job)
	}
	sort.Slice(workflow.Jobs, func(i, j int) bool { return workflow.Jobs[i].ID < workflow.Jobs[j].ID })
	return workflow, nil
}

func parseTriggers(node *yaml.Node) []string {
	seen := map[string]bool{}
	add := func(value string) {
		if value != "" {
			seen[value] = true
		}
	}
	switch {
	case node == nil:
	case node.Kind == yaml.ScalarNode:
		add(node.Value)
	case node.Kind == yaml.SequenceNode:
		for _, item := range node.Content {
			add(item.Value)
		}
	case node.Kind == yaml.MappingNode:
		for _, pair := range pairs(node) {
			add(pair.Key.Value)
		}
	}
	var out []string
	for trigger := range seen {
		out = append(out, trigger)
	}
	sort.Strings(out)
	return out
}

type pair struct {
	Key   *yaml.Node
	Value *yaml.Node
}

func pairs(node *yaml.Node) []pair {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	var out []pair
	for i := 0; i+1 < len(node.Content); i += 2 {
		out = append(out, pair{Key: node.Content[i], Value: node.Content[i+1]})
	}
	return out
}

func lookup(node *yaml.Node, key string) *yaml.Node {
	for _, pair := range pairs(node) {
		if pair.Key.Value == key {
			return pair.Value
		}
	}
	return nil
}

func scalar(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}
	return node.Value
}
