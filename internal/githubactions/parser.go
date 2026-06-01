package githubactions

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

func ParseWorkflow(file string, data []byte) (Workflow, error) {
	root, err := ParseYAMLDocument(data)
	if err != nil {
		return Workflow{}, fmt.Errorf("%s: %w", file, err)
	}
	workflow := Workflow{File: file, Root: root, Triggers: ParseTriggers(Lookup(root, "on"))}
	jobs := Lookup(root, "jobs")
	for _, pair := range Pairs(jobs) {
		if pair.Value.Kind != yaml.MappingNode {
			continue
		}
		job := Job{
			ID:          pair.Key.Value,
			Node:        pair.Value,
			Line:        pair.Value.Line,
			Permissions: ParsePermissions(Lookup(pair.Value, "permissions")),
		}
		steps := Lookup(pair.Value, "steps")
		if steps == nil || steps.Kind != yaml.SequenceNode {
			workflow.Jobs = append(workflow.Jobs, job)
			continue
		}
		for i, stepNode := range steps.Content {
			if stepNode.Kind != yaml.MappingNode {
				continue
			}
			job.Steps = append(job.Steps, Step{
				Index: i,
				Node:  stepNode,
				Name:  Scalar(Lookup(stepNode, "name")),
				ID:    Scalar(Lookup(stepNode, "id")),
				Uses:  Scalar(Lookup(stepNode, "uses")),
				Run:   Scalar(Lookup(stepNode, "run")),
				With:  MapToStringNodes(Lookup(stepNode, "with")),
				Line:  stepNode.Line,
			})
		}
		workflow.Jobs = append(workflow.Jobs, job)
	}
	return workflow, nil
}

func ParseTriggers(node *yaml.Node) []string {
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
		for _, pair := range Pairs(node) {
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

func TriggerSet(triggers []string) map[string]bool {
	out := map[string]bool{}
	for _, trigger := range triggers {
		out[trigger] = true
	}
	return out
}
