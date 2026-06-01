package guard

import (
	"fmt"
	"strings"
)

func extractFlagValue(text string, flag string) string {
	fields := strings.Fields(strings.NewReplacer("[", " ", "]", " ", ",", " ", `"`, " ", "'", " ").Replace(text))
	for i, field := range fields {
		if field == flag && i+1 < len(fields) {
			return fields[i+1]
		}
		if strings.HasPrefix(field, flag+"=") {
			return strings.TrimPrefix(field, flag+"=")
		}
	}
	return flag
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func stepLabel(step stepInfo) string {
	if step.name != "" {
		return step.name
	}
	if step.id != "" {
		return step.id
	}
	if step.uses != "" {
		return step.uses
	}
	return fmt.Sprintf("step %d", step.index+1)
}

func invocationLabel(inv *codexInvocation) string {
	return fmt.Sprintf("%s job %s at %s:%d", inv.summary.Kind, inv.job.id, inv.summary.File, inv.summary.Line)
}

func nonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 1
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
