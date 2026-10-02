package linter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Severity levels for query lint warnings.
type Severity string

const (
	SeverityInfo        Severity = "info"
	SeverityWarning     Severity = "warning"
	SeverityPerformance Severity = "performance"
)

// LintWarning represents a detected anti-pattern or performance tip.
type LintWarning struct {
	RuleID     string   `json:"rule_id"`
	Severity   Severity `json:"severity"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion"`
}

// AnalyzeQuery parses query JSON and detects common Elasticsearch anti-patterns.
func AnalyzeQuery(rawJSON string, fieldTypes map[string]string) []LintWarning {
	var warnings []LintWarning

	trimmed := strings.TrimSpace(rawJSON)
	if trimmed == "" {
		return warnings
	}

	// Extract JSON body if Kibana REST method header is present
	lines := strings.Split(trimmed, "\n")
	firstLine := strings.TrimSpace(lines[0])
	if isHTTPHeader(firstLine) {
		trimmed = strings.TrimSpace(strings.Join(lines[1:], "\n"))
		if trimmed == "" {
			return warnings
		}
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		// Not valid JSON or partial
		return warnings
	}

	// Rule 1: Check deep pagination (size / from)
	checkPagination(parsed, &warnings)

	// Rule 2: Check query clauses (must vs filter, leading wildcard)
	if queryMap, ok := parsed["query"].(map[string]interface{}); ok {
		checkQueryClauses(queryMap, &warnings)
	}

	// Rule 3: Check aggregations and sort against field types
	checkAggregations(parsed, fieldTypes, &warnings)
	checkSorting(parsed, fieldTypes, &warnings)

	return warnings
}

func isHTTPHeader(line string) bool {
	upper := strings.ToUpper(line)
	for _, m := range []string{"GET ", "POST ", "PUT ", "DELETE "} {
		if strings.HasPrefix(upper, m) {
			return true
		}
	}
	return strings.HasPrefix(line, "/")
}

func checkPagination(root map[string]interface{}, warnings *[]LintWarning) {
	var sizeVal float64 = 10
	var fromVal float64 = 0

	if s, ok := root["size"].(float64); ok {
		sizeVal = s
	}
	if f, ok := root["from"].(float64); ok {
		fromVal = f
	}

	if fromVal+sizeVal > 10000 {
		*warnings = append(*warnings, LintWarning{
			RuleID:     "MAX_RESULT_WINDOW_EXCEEDED",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("'from' (%d) + 'size' (%d) exceeds Elasticsearch's default index.max_result_window (10,000).", int(fromVal), int(sizeVal)),
			Suggestion: "Use 'search_after' parameter with a tie-breaker sort for efficient deep pagination without memory overhead.",
		})
	} else if sizeVal > 1000 {
		*warnings = append(*warnings, LintWarning{
			RuleID:     "LARGE_RESULT_SIZE",
			Severity:   SeverityPerformance,
			Message:    fmt.Sprintf("Requesting %d documents in a single request can cause high JVM heap pressure and network latency.", int(sizeVal)),
			Suggestion: "Keep 'size' <= 100 or use pagination ('from'/'size' or 'search_after') to fetch batches incrementally.",
		})
	}
}

func checkQueryClauses(queryMap map[string]interface{}, warnings *[]LintWarning) {
	// Inspect bool query
	if boolMap, ok := queryMap["bool"].(map[string]interface{}); ok {
		// Check if must contains term / terms / range queries that could be cached in filter
		if mustList, ok := boolMap["must"].([]interface{}); ok {
			for _, item := range mustList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					for qType := range itemMap {
						if qType == "term" || qType == "terms" || qType == "range" || qType == "exists" {
							*warnings = append(*warnings, LintWarning{
								RuleID:     "MUST_CAN_BE_FILTER",
								Severity:   SeverityPerformance,
								Message:    fmt.Sprintf("Clause '%s' is placed in 'must' context where relevance scoring is calculated.", qType),
								Suggestion: "Move exact-match queries ('term', 'terms', 'range', 'exists') to 'filter' context so Elasticsearch can automatically cache results and skip scoring.",
							})
							break
						}
					}
				}
			}
		}
	}

	// Recursively search for leading wildcards
	searchLeadingWildcard(queryMap, warnings)
}

func searchLeadingWildcard(node interface{}, warnings *[]LintWarning) {
	switch val := node.(type) {
	case map[string]interface{}:
		for k, v := range val {
			if k == "wildcard" {
				if fieldMap, ok := v.(map[string]interface{}); ok {
					for fName, patternVal := range fieldMap {
						pattern := fmt.Sprintf("%v", patternVal)
						if strings.HasPrefix(pattern, "*") || strings.HasPrefix(pattern, "?") {
							*warnings = append(*warnings, LintWarning{
								RuleID:     "LEADING_WILDCARD",
								Severity:   SeverityWarning,
								Message:    fmt.Sprintf("Leading wildcard '%s' on field '%s' forces a full index term scan.", pattern, fName),
								Suggestion: "Avoid starting wildcard patterns with '*' or '?'. Consider using an edge_ngram analyzer or prefix query instead.",
							})
						}
					}
				}
			} else {
				searchLeadingWildcard(v, warnings)
			}
		}
	case []interface{}:
		for _, item := range val {
			searchLeadingWildcard(item, warnings)
		}
	}
}

func checkAggregations(root map[string]interface{}, fieldTypes map[string]string, warnings *[]LintWarning) {
	aggs, ok := root["aggs"].(map[string]interface{})
	if !ok {
		aggs, ok = root["aggregations"].(map[string]interface{})
	}
	if !ok {
		return
	}

	if fieldTypes == nil {
		return
	}

	for _, aggDef := range aggs {
		aggMap, ok := aggDef.(map[string]interface{})
		if !ok {
			continue
		}
		if termsAgg, ok := aggMap["terms"].(map[string]interface{}); ok {
			if fieldName, ok := termsAgg["field"].(string); ok {
				if fType, exists := fieldTypes[fieldName]; exists && fType == "text" {
					*warnings = append(*warnings, LintWarning{
						RuleID:     "TEXT_FIELD_AGGREGATION",
						Severity:   SeverityWarning,
						Message:    fmt.Sprintf("Aggregation requested on 'text' field '%s'. Text fields cannot be aggregated without memory-heavy fielddata=true.", fieldName),
						Suggestion: fmt.Sprintf("Use the sub-field '%s.keyword' for aggregations instead of '%s'.", fieldName, fieldName),
					})
				}
			}
		}
	}
}

func checkSorting(root map[string]interface{}, fieldTypes map[string]string, warnings *[]LintWarning) {
	sortVal, ok := root["sort"]
	if !ok || fieldTypes == nil {
		return
	}

	checkSortField := func(field string) {
		if fType, exists := fieldTypes[field]; exists && fType == "text" {
			*warnings = append(*warnings, LintWarning{
				RuleID:     "TEXT_FIELD_SORT",
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("Sorting on 'text' field '%s' requires fielddata=true which consumes large JVM memory.", field),
				Suggestion: fmt.Sprintf("Sort on '%s.keyword' instead of '%s'.", field, field),
			})
		}
	}

	switch s := sortVal.(type) {
	case string:
		checkSortField(s)
	case []interface{}:
		for _, item := range s {
			if strField, ok := item.(string); ok {
				checkSortField(strField)
			} else if mapField, ok := item.(map[string]interface{}); ok {
				for k := range mapField {
					checkSortField(k)
				}
			}
		}
	}
}
