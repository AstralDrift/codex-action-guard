package guard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONReportSchemaValidatesReports(t *testing.T) {
	schema := readJSONSchema(t, "report.schema.json")
	for _, fixture := range []string{"secure", "vulnerable"} {
		t.Run(fixture, func(t *testing.T) {
			report, err := AuditPath(filepath.Join("..", "..", "fixtures", fixture), AuditOptions{All: true, ToolVersion: "test"})
			if err != nil {
				t.Fatal(err)
			}
			data, err := RenderJSON(report)
			if err != nil {
				t.Fatal(err)
			}
			validateJSONWithSchema(t, schema, data)
		})
	}
}

func TestRulesSchemaValidatesRulesExport(t *testing.T) {
	schema := readJSONSchema(t, "rules.schema.json")
	data, err := RenderRulesJSON("test")
	if err != nil {
		t.Fatal(err)
	}
	validateJSONWithSchema(t, schema, data)
}

func readJSONSchema(t *testing.T, name string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "schemas", name))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema %s is invalid JSON: %v", name, err)
	}
	return schema
}

func validateJSONWithSchema(t *testing.T, schema map[string]any, data []byte) {
	t.Helper()
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	if err := validateSchemaNode("$", schema, value); err != nil {
		t.Fatal(err)
	}
}

func validateSchemaNode(path string, schema map[string]any, value any) error {
	if enumValues, ok := schema["enum"].([]any); ok {
		var matched bool
		for _, enumValue := range enumValues {
			if value == enumValue {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%s: value %#v is not in enum %#v", path, value, enumValues)
		}
	}

	schemaType, _ := schema["type"].(string)
	switch schemaType {
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object, got %T", path, value)
		}
		required, err := schemaStringList(schema, "required")
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		for _, key := range required {
			if _, ok := object[key]; !ok {
				return fmt.Errorf("%s: missing required property %q", path, key)
			}
		}
		properties := schemaProperties(schema)
		if additional, ok := schema["additionalProperties"].(bool); ok && !additional {
			for key := range object {
				if _, known := properties[key]; !known {
					return fmt.Errorf("%s: unexpected property %q", path, key)
				}
			}
		}
		for key, child := range properties {
			childValue, ok := object[key]
			if !ok {
				continue
			}
			if err := validateSchemaNode(path+"."+key, child, childValue); err != nil {
				return err
			}
		}
	case "array":
		array, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s: expected array, got %T", path, value)
		}
		items, ok := schema["items"].(map[string]any)
		if !ok {
			return nil
		}
		for i, item := range array {
			if err := validateSchemaNode(fmt.Sprintf("%s[%d]", path, i), items, item); err != nil {
				return err
			}
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s: expected string, got %T", path, value)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%s: expected number, got %T", path, value)
		}
	case "":
		return nil
	default:
		return fmt.Errorf("%s: unsupported schema type %q", path, schemaType)
	}
	return nil
}

func schemaStringList(schema map[string]any, key string) ([]string, error) {
	raw, ok := schema[key]
	if !ok {
		return nil, nil
	}
	values, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array", key)
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%s contains non-string value %T", key, value)
		}
		out = append(out, text)
	}
	return out, nil
}

func schemaProperties(schema map[string]any) map[string]map[string]any {
	out := map[string]map[string]any{}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return out
	}
	for key, raw := range properties {
		child, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		out[strings.TrimSpace(key)] = child
	}
	return out
}
