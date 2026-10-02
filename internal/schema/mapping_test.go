package schema

import (
	"encoding/json"
	"testing"
)

func TestParseMappingJSON_ES7_ES8(t *testing.T) {
	// Standard ES 7.x/8.x mapping JSON
	rawJSON := `{
		"users": {
			"mappings": {
				"properties": {
					"id": { "type": "long" },
					"name": {
						"type": "text",
						"fields": {
							"keyword": { "type": "keyword", "ignore_above": 256 }
						}
					},
					"profile": {
						"properties": {
							"age": { "type": "integer" },
							"city": { "type": "keyword" }
						}
					},
					"created_at": { "type": "date" }
				}
			}
		}
	}`

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &raw); err != nil {
		t.Fatalf("failed to unmarshal test JSON: %v", err)
	}

	fields := ParseMappingJSON(raw, "users")
	fieldMap := make(map[string]FieldSuggestion)
	for _, f := range fields {
		fieldMap[f.Name] = f
	}

	// Verify standard fields
	if _, ok := fieldMap["_id"]; !ok {
		t.Errorf("expected _id in suggestions")
	}

	// Verify primary fields
	if f, ok := fieldMap["id"]; !ok || f.Type != "long" {
		t.Errorf("expected id of type long, got %+v", f)
	}
	if f, ok := fieldMap["name"]; !ok || f.Type != "text" {
		t.Errorf("expected name of type text, got %+v", f)
	}

	// Verify multi-field (.keyword)
	if f, ok := fieldMap["name.keyword"]; !ok || f.Type != "keyword" {
		t.Errorf("expected name.keyword of type keyword, got %+v", f)
	}

	// Verify nested object dot notation
	if f, ok := fieldMap["profile.age"]; !ok || f.Type != "integer" {
		t.Errorf("expected profile.age of type integer, got %+v", f)
	}
	if f, ok := fieldMap["profile.city"]; !ok || f.Type != "keyword" {
		t.Errorf("expected profile.city of type keyword, got %+v", f)
	}
}

func TestParseMappingJSON_TypeWrapped(t *testing.T) {
	// Legacy type wrapper (e.g. "_doc")
	rawJSON := `{
		"logs": {
			"mappings": {
				"_doc": {
					"properties": {
						"level": { "type": "keyword" },
						"message": { "type": "text" }
					}
				}
			}
		}
	}`

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	fields := ParseMappingJSON(raw, "logs")
	fieldMap := make(map[string]FieldSuggestion)
	for _, f := range fields {
		fieldMap[f.Name] = f
	}

	if f, ok := fieldMap["level"]; !ok || f.Type != "keyword" {
		t.Errorf("expected level of type keyword, got %+v", f)
	}
	if f, ok := fieldMap["message"]; !ok || f.Type != "text" {
		t.Errorf("expected message of type text, got %+v", f)
	}
}

func TestGetStandardSnippets(t *testing.T) {
	snippets := GetStandardSnippets()
	if len(snippets) == 0 {
		t.Errorf("expected non-empty snippets list")
	}
}
