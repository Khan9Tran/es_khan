package linter

import (
	"testing"
)

func TestAnalyzeQuery_AntiPatterns(t *testing.T) {
	fieldTypes := map[string]string{
		"title":       "text",
		"status":      "keyword",
		"created_at":  "date",
	}

	// 1. Test Leading Wildcard
	queryWithWildcard := `{
		"query": {
			"wildcard": {
				"title": "*test"
			}
		}
	}`
	warnings := AnalyzeQuery(queryWithWildcard, fieldTypes)
	hasWildcardWarning := false
	for _, w := range warnings {
		if w.RuleID == "LEADING_WILDCARD" {
			hasWildcardWarning = true
			break
		}
	}
	if !hasWildcardWarning {
		t.Errorf("expected LEADING_WILDCARD warning, got none")
	}

	// 2. Test Must with Term query (should be filter)
	queryMustTerm := `{
		"query": {
			"bool": {
				"must": [
					{ "term": { "status": "active" } }
				]
			}
		}
	}`
	warningsMust := AnalyzeQuery(queryMustTerm, fieldTypes)
	hasMustWarning := false
	for _, w := range warningsMust {
		if w.RuleID == "MUST_CAN_BE_FILTER" {
			hasMustWarning = true
			break
		}
	}
	if !hasMustWarning {
		t.Errorf("expected MUST_CAN_BE_FILTER warning, got none")
	}

	// 3. Test Text field aggregation
	queryTextAgg := `{
		"aggs": {
			"by_title": {
				"terms": { "field": "title" }
			}
		}
	}`
	warningsAgg := AnalyzeQuery(queryTextAgg, fieldTypes)
	hasAggWarning := false
	for _, w := range warningsAgg {
		if w.RuleID == "TEXT_FIELD_AGGREGATION" {
			hasAggWarning = true
			break
		}
	}
	if !hasAggWarning {
		t.Errorf("expected TEXT_FIELD_AGGREGATION warning, got none")
	}

	// 4. Test Large Result Size
	queryLargeSize := `{
		"size": 5000,
		"query": { "match_all": {} }
	}`
	warningsSize := AnalyzeQuery(queryLargeSize, fieldTypes)
	hasSizeWarning := false
	for _, w := range warningsSize {
		if w.RuleID == "LARGE_RESULT_SIZE" {
			hasSizeWarning = true
			break
		}
	}
	if !hasSizeWarning {
		t.Errorf("expected LARGE_RESULT_SIZE warning, got none")
	}
}
