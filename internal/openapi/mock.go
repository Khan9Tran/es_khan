package openapi

import (
	"encoding/json"
	"strings"
)

// GenerateMockFromSchema generates an indented JSON string from a schema definition and root spec doc.
func GenerateMockFromSchema(schema map[string]any, rootDoc map[string]any) string {
	if len(schema) == 0 {
		return "{}"
	}

	visited := make(map[string]int)
	obj := buildMock(schema, rootDoc, visited, 0)
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func buildMock(schema map[string]any, rootDoc map[string]any, visited map[string]int, depth int) any {
	if depth > 6 {
		return map[string]any{}
	}

	// Resolve $ref
	if refVal, ok := schema["$ref"].(string); ok && refVal != "" {
		if visited[refVal] > 1 {
			return map[string]any{}
		}
		visited[refVal]++
		defer func() { visited[refVal]-- }()

		resolved := resolveRef(refVal, rootDoc)
		if resolved != nil {
			return buildMock(resolved, rootDoc, visited, depth+1)
		}
		return map[string]any{}
	}

	// If enum exists, return first item
	if enumVals, ok := schema["enum"].([]any); ok && len(enumVals) > 0 {
		return enumVals[0]
	}

	// If default value exists, prefer it
	if defVal, ok := schema["default"]; ok && defVal != nil {
		return defVal
	}

	schemaType := ""
	if t, ok := schema["type"].(string); ok {
		schemaType = strings.ToLower(t)
	}

	// If type is object or properties exist
	if props, ok := schema["properties"].(map[string]any); ok {
		result := make(map[string]any)
		for propName, propVal := range props {
			if propSchema, ok := propVal.(map[string]any); ok {
				result[propName] = buildMock(propSchema, rootDoc, visited, depth+1)
			} else {
				result[propName] = ""
			}
		}
		return result
	}

	// If type is array
	if schemaType == "array" || schema["items"] != nil {
		if itemsSchema, ok := schema["items"].(map[string]any); ok {
			return []any{buildMock(itemsSchema, rootDoc, visited, depth+1)}
		}
		return []any{}
	}

	// Primitive types
	switch schemaType {
	case "string":
		format, _ := schema["format"].(string)
		switch format {
		case "date-time":
			return "2026-10-02T12:00:00Z"
		case "date":
			return "2026-10-02"
		case "email":
			return "user@example.com"
		case "uuid":
			return "3fa85f64-5717-4562-b3fc-2c963f66afa6"
		default:
			if ex, ok := schema["example"].(string); ok && ex != "" {
				return ex
			}
			return "string"
		}

	case "integer":
		if ex, ok := schema["example"]; ok {
			return ex
		}
		return 0

	case "number":
		if ex, ok := schema["example"]; ok {
			return ex
		}
		return 0.0

	case "boolean":
		if ex, ok := schema["example"]; ok {
			return ex
		}
		return true

	case "object":
		return map[string]any{}

	default:
		return ""
	}
}

func resolveRef(ref string, rootDoc map[string]any) map[string]any {
	if !strings.HasPrefix(ref, "#/") {
		return nil
	}

	parts := strings.Split(ref[2:], "/")
	var current any = rootDoc

	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		val, exists := m[part]
		if !exists {
			return nil
		}
		current = val
	}

	if result, ok := current.(map[string]any); ok {
		return result
	}
	return nil
}
