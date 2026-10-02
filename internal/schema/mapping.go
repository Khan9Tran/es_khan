package schema

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"eskhan/internal/es"
)

// FieldSuggestion contains autocomplete information for a field in an index.
type FieldSuggestion struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Detail      string `json:"detail"`
	Description string `json:"description,omitempty"`
}

// IndexSchema contains the analyzed fields and metadata for an index.
type IndexSchema struct {
	IndexName string            `json:"index_name"`
	Fields    []FieldSuggestion `json:"fields"`
}

// Manager maintains in-memory cached schemas for index autocomplete.
type Manager struct {
	mu     sync.RWMutex
	cache  map[string]IndexSchema
	client *es.Client
}

// NewManager creates a schema manager.
func NewManager(client *es.Client) *Manager {
	return &Manager{
		cache:  make(map[string]IndexSchema),
		client: client,
	}
}

// UpdateClient updates the ES client used for fetching mappings.
func (m *Manager) UpdateClient(client *es.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.client = client
	m.cache = make(map[string]IndexSchema) // Invalidate cache on client switch
}

// GetIndexFields retrieves the field suggestions for an index, using cache when available.
func (m *Manager) GetIndexFields(ctx context.Context, indexName string) ([]FieldSuggestion, error) {
	if indexName == "" {
		return DefaultESFields(), nil
	}

	m.mu.RLock()
	cached, found := m.cache[indexName]
	m.mu.RUnlock()
	if found {
		return cached.Fields, nil
	}

	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("no active elasticsearch client configured")
	}

	rawMapping, err := client.GetMapping(ctx, indexName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mapping for %s: %w", indexName, err)
	}

	fields := ParseMappingJSON(rawMapping, indexName)

	m.mu.Lock()
	m.cache[indexName] = IndexSchema{
		IndexName: indexName,
		Fields:    fields,
	}
	m.mu.Unlock()

	return fields, nil
}

// ParseMappingJSON parses the response from /{index}/_mapping into field suggestions.
// Supports both ES 7.x/8.x (no type) and older ES with type wrapper (e.g. "_doc").
func ParseMappingJSON(raw map[string]interface{}, targetIndex string) []FieldSuggestion {
	var results []FieldSuggestion
	fieldsMap := make(map[string]FieldSuggestion)

	// Built-in metadata fields always available
	for _, meta := range DefaultESFields() {
		fieldsMap[meta.Name] = meta
	}

	for idxName, idxData := range raw {
		if targetIndex != "" && targetIndex != "*" && idxName != targetIndex {
			continue
		}

		idxMap, ok := idxData.(map[string]interface{})
		if !ok {
			continue
		}

		mappings, ok := idxMap["mappings"].(map[string]interface{})
		if !ok {
			continue
		}

		// Check if "properties" is at top-level of mappings (ES 7.x+)
		var properties map[string]interface{}
		if props, ok := mappings["properties"].(map[string]interface{}); ok {
			properties = props
		} else {
			// Check if nested under type (e.g. mappings["_doc"]["properties"])
			for _, v := range mappings {
				if typeMap, ok := v.(map[string]interface{}); ok {
					if props, ok := typeMap["properties"].(map[string]interface{}); ok {
						properties = props
						break
					}
				}
			}
		}

		if properties != nil {
			extractProperties("", properties, fieldsMap)
		}
	}

	for _, f := range fieldsMap {
		results = append(results, f)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// extractProperties recursively extracts fields, handling nested objects and multi-fields.
func extractProperties(prefix string, props map[string]interface{}, out map[string]FieldSuggestion) {
	for fieldName, fieldVal := range props {
		fieldMap, ok := fieldVal.(map[string]interface{})
		if !ok {
			continue
		}

		fullPath := fieldName
		if prefix != "" {
			fullPath = prefix + "." + fieldName
		}

		fieldType, _ := fieldMap["type"].(string)
		if fieldType == "" {
			// If no explicit type, check if it's an object with nested properties
			if subProps, ok := fieldMap["properties"].(map[string]interface{}); ok {
				fieldType = "object"
				out[fullPath] = FieldSuggestion{
					Name:   fullPath,
					Type:   "object",
					Detail: "Object containing sub-fields",
				}
				extractProperties(fullPath, subProps, out)
				continue
			}
		}

		out[fullPath] = FieldSuggestion{
			Name:   fullPath,
			Type:   fieldType,
			Detail: fmt.Sprintf("[%s] Field in mapping", fieldType),
		}

		// Handle sub-properties for nested/object types
		if subProps, ok := fieldMap["properties"].(map[string]interface{}); ok {
			extractProperties(fullPath, subProps, out)
		}

		// Handle multi-fields (e.g. text field with "fields": { "keyword": { "type": "keyword" } })
		if subFields, ok := fieldMap["fields"].(map[string]interface{}); ok {
			for subName, subVal := range subFields {
				if subMap, ok := subVal.(map[string]interface{}); ok {
					subType, _ := subMap["type"].(string)
					subFullPath := fullPath + "." + subName
					out[subFullPath] = FieldSuggestion{
						Name:   subFullPath,
						Type:   subType,
						Detail: fmt.Sprintf("[%s] Multi-field of %s", subType, fullPath),
					}
				}
			}
		}
	}
}

// DefaultESFields returns metadata fields present in all Elasticsearch indices.
func DefaultESFields() []FieldSuggestion {
	return []FieldSuggestion{
		{Name: "_id", Type: "keyword", Detail: "Document unique identifier"},
		{Name: "_index", Type: "keyword", Detail: "Index to which document belongs"},
		{Name: "_score", Type: "float", Detail: "Relevance score"},
		{Name: "_source", Type: "object", Detail: "Original JSON document body"},
	}
}

// DSLSnippet represents an autocomplete snippet for Elasticsearch Query DSL.
type DSLSnippet struct {
	Label       string `json:"label"`
	InsertText  string `json:"insert_text"`
	Kind        string `json:"kind"`
	Detail      string `json:"detail"`
	Doc         string `json:"documentation"`
}

// GetStandardSnippets returns pre-configured snippets for Monaco Editor.
func GetStandardSnippets() []DSLSnippet {
	return []DSLSnippet{
		{
			Label:      "bool query",
			InsertText: "{\n  \"query\": {\n    \"bool\": {\n      \"must\": [\n        { \"match\": { \"${1:field}\": \"${2:value}\" } }\n      ],\n      \"filter\": [\n        { \"term\": { \"${3:status}\": \"${4:active}\" } }\n      ]\n    }\n  }\n}",
			Kind:       "Snippet",
			Detail:     "Bool query with must and filter",
			Doc:        "Matches documents matching boolean combinations of other queries.",
		},
		{
			Label:      "match_all",
			InsertText: "{\n  \"query\": {\n    \"match_all\": {}\n  }\n}",
			Kind:       "Snippet",
			Detail:     "Match all documents",
			Doc:        "The most simple query, which matches all documents.",
		},
		{
			Label:      "range query",
			InsertText: "{\n  \"range\": {\n    \"${1:field}\": {\n      \"gte\": ${2:10},\n      \"lte\": ${3:100}\n    }\n  }\n}",
			Kind:       "Snippet",
			Detail:     "Range query (numbers, dates)",
			Doc:        "Matches documents with terms within a provided range.",
		},
		{
			Label:      "terms aggregation",
			InsertText: "{\n  \"size\": 0,\n  \"aggs\": {\n    \"${1:by_category}\": {\n      \"terms\": {\n        \"field\": \"${2:category.keyword}\",\n        \"size\": ${3:10}\n      }\n    }\n  }\n}",
			Kind:       "Snippet",
			Detail:     "Terms aggregation",
			Doc:        "Multi-bucket value aggregation where buckets are dynamically built.",
		},
		{
			Label:      "date_histogram aggregation",
			InsertText: "{\n  \"size\": 0,\n  \"aggs\": {\n    \"${1:over_time}\": {\n      \"date_histogram\": {\n        \"field\": \"${2:@timestamp}\",\n        \"calendar_interval\": \"${3:day}\"\n      }\n    }\n  }\n}",
			Kind:       "Snippet",
			Detail:     "Date histogram aggregation",
			Doc:        "Multi-bucket aggregation over date values.",
		},
		{
			Label:      "highlight",
			InsertText: "\"highlight\": {\n  \"pre_tags\": [\"<mark>\"],\n  \"post_tags\": [\"</mark>\"],\n  \"fields\": {\n    \"${1:content}\": {}\n  }\n}",
			Kind:       "Snippet",
			Detail:     "Search hit highlight with <mark> tags",
			Doc:        "Highlights search matches in query results.",
		},
		{
			Label:      "vietnamese multi_match",
			InsertText: "{\n  \"query\": {\n    \"multi_match\": {\n      \"query\": \"${1:điện thoại thông minh}\",\n      \"fields\": [\"${2:name^3}\", \"${3:title^2}\", \"${4:description}\"],\n      \"type\": \"best_fields\",\n      \"operator\": \"and\"\n    }\n  },\n  \"highlight\": {\n    \"pre_tags\": [\"<mark>\"],\n    \"post_tags\": [\"</mark>\"],\n    \"fields\": {\n      \"${2:name}\": {},\n      \"${4:description}\": {}\n    }\n  }\n}",
			Kind:       "Snippet",
			Detail:     "Vietnamese multi-match search with field weights & highlight",
			Doc:        "Optimized full-text search for Vietnamese text (accents/diacritics, weighted fields, and matching highlights).",
		},
		{
			Label:      "vietnamese phrase_prefix",
			InsertText: "{\n  \"query\": {\n    \"match_phrase_prefix\": {\n      \"${1:name}\": {\n        \"query\": \"${2:máy giặt}\",\n        \"max_expansions\": ${3:10}\n      }\n    }\n  }\n}",
			Kind:       "Snippet",
			Detail:     "Vietnamese search-as-you-type autocomplete",
			Doc:        "Instant autocomplete / search suggestion for Vietnamese queries as user types.",
		},
		{
			Label:      "vietnamese bool_filter",
			InsertText: "{\n  \"query\": {\n    \"bool\": {\n      \"must\": [\n        {\n          \"multi_match\": {\n            \"query\": \"${1:áo sơ mi}\",\n            \"fields\": [\"${2:title^2}\", \"${3:content}\"],\n            \"operator\": \"and\"\n          }\n        }\n      ],\n      \"filter\": [\n        { \"term\": { \"${4:status}\": \"${5:active}\" } },\n        { \"range\": { \"${6:price}\": { \"gte\": ${7:100000}, \"lte\": ${8:500000} } } }\n      ]\n    }\n  },\n  \"sort\": [\n    { \"_score\": { \"order\": \"desc\" } }\n  ]\n}",
			Kind:       "Snippet",
			Detail:     "Vietnamese query with filters & scoring",
			Doc:        "Compound query combining Vietnamese full-text search with exact terms and range filters.",
		},
	}
}
