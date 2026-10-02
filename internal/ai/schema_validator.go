package ai

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"eskhan/internal/schema"
)

// ClarificationOption represents an interactive suggestion for the user when a field is uncertain or invalid.
type ClarificationOption struct {
	Label string `json:"label"`
	Field string `json:"field"`
	Type  string `json:"type,omitempty"`
}

// FieldValidationResult holds validation outcome of a query against an index schema.
type FieldValidationResult struct {
	IsValid          bool                  `json:"is_valid"`
	VerifiedFields   []string              `json:"verified_fields"`
	InvalidFields    []string              `json:"invalid_fields"`
	Question         string                `json:"question,omitempty"`
	SuggestedOptions []ClarificationOption `json:"suggested_options,omitempty"`
}

var esMetadataFields = map[string]bool{
	"_id":           true,
	"_score":        true,
	"_index":        true,
	"_source":       true,
	"_doc":          true,
	"_type":         true,
	"_version":      true,
	"_seq_no":       true,
	"_primary_term": true,
	"_routing":      true,
	"_shard":        true,
	"_node":         true,
}

// ExtractQueryFields parses an Elasticsearch JSON query and extracts all referenced field names.
func ExtractQueryFields(rawQuery string) ([]string, error) {
	cleanJSON := cleanJSONContent(rawQuery)
	if cleanJSON == "" {
		return nil, fmt.Errorf("empty query content")
	}

	var root interface{}
	if err := json.Unmarshal([]byte(cleanJSON), &root); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	fieldSet := make(map[string]bool)
	walkForFields(root, fieldSet)

	var list []string
	for f := range fieldSet {
		if f != "" {
			list = append(list, f)
		}
	}
	sort.Strings(list)
	return list, nil
}

func cleanJSONContent(raw string) string {
	trimmed := strings.TrimSpace(raw)
	lines := strings.Split(trimmed, "\n")
	if len(lines) == 0 {
		return ""
	}
	firstLine := strings.TrimSpace(lines[0])
	if regexpMethodMatch.MatchString(firstLine) || strings.HasPrefix(firstLine, "/") {
		return strings.TrimSpace(strings.Join(lines[1:], "\n"))
	}
	return trimmed
}

var regexpMethodMatch = regexp.MustCompile(`^(GET|POST|PUT|DELETE)\s+`)

func walkForFields(node interface{}, fieldSet map[string]bool) {
	switch v := node.(type) {
	case map[string]interface{}:
		for k, val := range v {
			switch k {
			case "term", "terms", "match", "match_phrase", "match_phrase_prefix", "match_bool_prefix",
				"range", "wildcard", "prefix", "fuzzy", "regexp":
				if subMap, ok := val.(map[string]interface{}); ok {
					for fieldName := range subMap {
						addField(fieldSet, fieldName)
					}
				}
			case "multi_match":
				if mm, ok := val.(map[string]interface{}); ok {
					if fields, ok := mm["fields"].([]interface{}); ok {
						for _, f := range fields {
							if str, ok := f.(string); ok {
								addField(fieldSet, str)
							}
						}
					}
				}
			case "exists":
				if ex, ok := val.(map[string]interface{}); ok {
					if f, ok := ex["field"].(string); ok {
						addField(fieldSet, f)
					}
				}
			case "nested":
				if n, ok := val.(map[string]interface{}); ok {
					if path, ok := n["path"].(string); ok {
						addField(fieldSet, path)
					}
					if q, ok := n["query"]; ok {
						walkForFields(q, fieldSet)
					}
				}
			case "sort":
				extractSortFields(val, fieldSet)
			case "aggs", "aggregations":
				extractAggsFields(val, fieldSet)
			case "highlight":
				if hl, ok := val.(map[string]interface{}); ok {
					if hlFields, ok := hl["fields"].(map[string]interface{}); ok {
						for f := range hlFields {
							addField(fieldSet, f)
						}
					}
				}
			case "_source":
				extractSourceFields(val, fieldSet)
			default:
				// Continue recursion for bool, must, should, filter, must_not, query, etc.
				walkForFields(val, fieldSet)
			}
		}
	case []interface{}:
		for _, item := range v {
			walkForFields(item, fieldSet)
		}
	}
}

func addField(fieldSet map[string]bool, field string) {
	cleaned := strings.TrimSpace(field)
	if cleaned == "" {
		return
	}
	// Strip boost e.g. "title^3" -> "title"
	if idx := strings.Index(cleaned, "^"); idx != -1 {
		cleaned = cleaned[:idx]
	}
	fieldSet[cleaned] = true
}

func extractSortFields(val interface{}, fieldSet map[string]bool) {
	switch s := val.(type) {
	case string:
		addField(fieldSet, s)
	case []interface{}:
		for _, item := range s {
			extractSortFields(item, fieldSet)
		}
	case map[string]interface{}:
		for k := range s {
			if k != "_geo_distance" && k != "_script" {
				addField(fieldSet, k)
			}
		}
	}
}

func extractAggsFields(val interface{}, fieldSet map[string]bool) {
	aggsMap, ok := val.(map[string]interface{})
	if !ok {
		return
	}
	for _, aggBody := range aggsMap {
		if bodyMap, ok := aggBody.(map[string]interface{}); ok {
			for _, aggTypeVal := range bodyMap {
				if innerMap, ok := aggTypeVal.(map[string]interface{}); ok {
					if f, ok := innerMap["field"].(string); ok {
						addField(fieldSet, f)
					}
				}
			}
			if subAggs, ok := bodyMap["aggs"]; ok {
				extractAggsFields(subAggs, fieldSet)
			}
			if subAggs, ok := bodyMap["aggregations"]; ok {
				extractAggsFields(subAggs, fieldSet)
			}
		}
	}
}

func extractSourceFields(val interface{}, fieldSet map[string]bool) {
	switch s := val.(type) {
	case string:
		addField(fieldSet, s)
	case []interface{}:
		for _, item := range s {
			if str, ok := item.(string); ok {
				addField(fieldSet, str)
			}
		}
	case map[string]interface{}:
		if inc, ok := s["includes"].([]interface{}); ok {
			for _, item := range inc {
				if str, ok := item.(string); ok {
					addField(fieldSet, str)
				}
			}
		}
	}
}

// ValidateQueryAgainstSchema inspects all fields referenced in query against known schema fields.
func ValidateQueryAgainstSchema(rawQuery string, indexName string, schemaFields []schema.FieldSuggestion) FieldValidationResult {
	fieldsInQuery, err := ExtractQueryFields(rawQuery)
	if err != nil {
		return FieldValidationResult{
			IsValid: true, // Cannot parse fields, fall back to letting ES handle execution
		}
	}

	if len(schemaFields) == 0 {
		return FieldValidationResult{
			IsValid:        true,
			VerifiedFields: fieldsInQuery,
		}
	}

	schemaMap := make(map[string]schema.FieldSuggestion, len(schemaFields))
	for _, f := range schemaFields {
		schemaMap[f.Name] = f
	}

	var verified []string
	var invalid []string

	for _, field := range fieldsInQuery {
		if esMetadataFields[field] {
			verified = append(verified, field)
			continue
		}

		// Exact match
		if _, ok := schemaMap[field]; ok {
			verified = append(verified, field)
			continue
		}

		// Multi-field check: e.g. "customer_name.keyword"
		if strings.HasSuffix(field, ".keyword") {
			base := strings.TrimSuffix(field, ".keyword")
			if sf, ok := schemaMap[base]; ok && (sf.Type == "text" || sf.Type == "keyword") {
				verified = append(verified, field)
				continue
			}
		}

		// Prefix check for object / nested parent fields (e.g. "customer" when schema has "customer.name")
		isObjectParent := false
		for sName := range schemaMap {
			if strings.HasPrefix(sName, field+".") {
				isObjectParent = true
				break
			}
		}
		if isObjectParent {
			verified = append(verified, field)
			continue
		}

		invalid = append(invalid, field)
	}

	if len(invalid) == 0 {
		return FieldValidationResult{
			IsValid:        true,
			VerifiedFields: verified,
		}
	}

	// Build interactive clarification question and suggested options
	var options []ClarificationOption
	seenOptionFields := make(map[string]bool)

	for _, invField := range invalid {
		candidates := FindClosestFields(invField, schemaFields, 3)
		for _, cand := range candidates {
			if !seenOptionFields[cand.Name] {
				seenOptionFields[cand.Name] = true
				options = append(options, ClarificationOption{
					Label: fmt.Sprintf("Dùng '%s' (%s) thay cho '%s'", cand.Name, cand.Type, invField),
					Field: cand.Name,
					Type:  cand.Type,
				})
			}
		}
	}

	targetIndexLabel := indexName
	if targetIndexLabel == "" {
		targetIndexLabel = "chỉ mục"
	}

	question := fmt.Sprintf(
		"Trong schema của '%s' không tồn tại trường '%s'. Bạn có muốn thay thế bằng một trong các trường gợi ý sau không?",
		targetIndexLabel,
		strings.Join(invalid, "', '"),
	)

	return FieldValidationResult{
		IsValid:          false,
		VerifiedFields:   verified,
		InvalidFields:    invalid,
		Question:         question,
		SuggestedOptions: options,
	}
}

// Synonym groups for concept matching between user intentions and schema field names.
var synonymGroups = [][]string{
	// Price & Money
	{"price", "amount", "cost", "total", "fee", "val", "gia", "tien", "sotien", "thanhtien", "unit_price", "total_amount"},
	// Status & State
	{"status", "state", "condition", "step", "trangthai", "tinhtrang", "order_status", "payment_status"},
	// Name & Title
	{"name", "title", "label", "text", "ten", "tieude", "customer_name", "product_name", "full_name"},
	// Date & Time
	{"date", "time", "created", "updated", "at", "timestamp", "ngay", "thoigian", "ngaytao", "created_at", "updated_at"},
	// User & Customer
	{"user", "customer", "creator", "account", "author", "client", "nguoidung", "khachhang", "khach", "user_id", "customer_id"},
	// Category & Type
	{"category", "type", "group", "tag", "genre", "loai", "danhmuc", "nhom", "category_id", "category_name"},
	// Code & ID
	{"code", "id", "sku", "uuid", "no", "number", "ma", "so", "order_code", "product_code"},
	// Quantity & Count
	{"count", "quantity", "qty", "soluong", "total_count", "item_count"},
	// Address & Location
	{"address", "location", "city", "country", "street", "diachi", "thanhpho", "shipping_address"},
	// Phone & Contact
	{"phone", "mobile", "tel", "sdt", "dienthoai", "phone_number"},
	// Email
	{"email", "mail", "email_address"},
}

// FindClosestFields searches for the best matching schema fields for an unknown/invalid field.
func FindClosestFields(targetField string, schemaFields []schema.FieldSuggestion, maxCount int) []schema.FieldSuggestion {
	if len(schemaFields) == 0 {
		return nil
	}
	targetLower := strings.ToLower(targetField)

	type scoredItem struct {
		field schema.FieldSuggestion
		score int
	}

	var scored []scoredItem

	for _, sf := range schemaFields {
		if strings.HasPrefix(sf.Name, "_") {
			continue // Skip built-in metadata
		}
		sLower := strings.ToLower(sf.Name)
		score := 0

		// 1. Synonym group check (+50)
		for _, group := range synonymGroups {
			targetInGroup := false
			schemaInGroup := false
			for _, syn := range group {
				if strings.Contains(targetLower, syn) || targetLower == syn {
					targetInGroup = true
				}
				if strings.Contains(sLower, syn) || sLower == syn {
					schemaInGroup = true
				}
			}
			if targetInGroup && schemaInGroup {
				score += 50
				break
			}
		}

		// 2. Substring match (+30)
		if strings.Contains(sLower, targetLower) || strings.Contains(targetLower, sLower) {
			score += 30
		}

		// 3. Normalized Levenshtein distance (+0 to 30)
		dist := levenshtein(targetLower, sLower)
		maxLen := int(math.Max(float64(len(targetLower)), float64(len(sLower))))
		if maxLen > 0 {
			sim := 1.0 - (float64(dist) / float64(maxLen))
			if sim > 0.3 {
				score += int(sim * 30)
			}
		}

		if score > 0 {
			scored = append(scored, scoredItem{field: sf, score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	var results []schema.FieldSuggestion
	for i := 0; i < len(scored) && i < maxCount; i++ {
		results = append(results, scored[i].field)
	}

	// Fallback if no matching score: return first maxCount fields from schema
	if len(results) == 0 {
		count := 0
		for _, sf := range schemaFields {
			if !strings.HasPrefix(sf.Name, "_") {
				results = append(results, sf)
				count++
				if count >= maxCount {
					break
				}
			}
		}
	}

	return results
}

func levenshtein(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
		dp[i][0] = i
	}
	for j := 0; j <= m; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 1
			if r1[i-1] == r2[j-1] {
				cost = 0
			}
			dp[i][j] = min(
				dp[i-1][j]+1,      // deletion
				dp[i][j-1]+1,      // insertion
				dp[i-1][j-1]+cost, // substitution
			)
		}
	}

	return dp[n][m]
}

func min(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= a && b <= c {
		return b
	}
	return c
}
