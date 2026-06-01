package githubactions

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type NodePair struct {
	Key   *yaml.Node
	Value *yaml.Node
}

func ParseYAMLDocument(data []byte) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("empty yaml document")
	}
	return doc.Content[0], nil
}

func Pairs(node *yaml.Node) []NodePair {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	out := make([]NodePair, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		out = append(out, NodePair{Key: node.Content[i], Value: node.Content[i+1]})
	}
	return out
}

func Lookup(node *yaml.Node, key string) *yaml.Node {
	for _, pair := range Pairs(node) {
		if pair.Key.Value == key {
			return pair.Value
		}
	}
	return nil
}

func Scalar(node *yaml.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == yaml.ScalarNode {
		return node.Value
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	_ = enc.Encode(node)
	_ = enc.Close()
	return strings.TrimSpace(buf.String())
}

func NodeText(node *yaml.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case yaml.ScalarNode:
		return node.Value
	case yaml.SequenceNode, yaml.MappingNode:
		var parts []string
		var walk func(*yaml.Node)
		walk = func(n *yaml.Node) {
			if n == nil {
				return
			}
			if n.Kind == yaml.ScalarNode && n.Value != "" {
				parts = append(parts, n.Value)
			}
			for _, child := range n.Content {
				walk(child)
			}
		}
		walk(node)
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func MapToStringNodes(node *yaml.Node) map[string]*yaml.Node {
	out := map[string]*yaml.Node{}
	for _, pair := range Pairs(node) {
		out[pair.Key.Value] = pair.Value
	}
	return out
}

func LineSnippet(lines []string, line int) string {
	if line <= 0 || line > len(lines) {
		return ""
	}
	return strings.TrimSpace(lines[line-1])
}

func FirstLineOf(text string, needle string, fallback int) int {
	if needle == "" {
		return fallback
	}
	idx := strings.Index(text, needle)
	if idx < 0 {
		return fallback
	}
	return strings.Count(text[:idx], "\n") + 1
}
