package converter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestQueryToGolang_MatchAll(t *testing.T) {
	c := NewConverter(nil)

	jsonInput := `{
		"query": {
			"match_all": {}
		},
		"from": 20,
		"size": 10
	}`

	code, err := c.QueryToGolang(jsonInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "QueryMatchAll()") {
		t.Errorf("expected code to contain QueryMatchAll(), got:\n%s", code)
	}
	if !strings.Contains(code, ".Paging(2, 10)") {
		t.Errorf("expected code to contain .Paging(2, 10), got:\n%s", code)
	}
}

func TestQueryToGolang_BoolFilters(t *testing.T) {
	c := NewConverter(nil)

	jsonInput := `POST /ecommerce_orders/_search
	{
		"query": {
			"bool": {
				"must": [
					{ "term": { "status": "completed" } },
					{ "match": { "customer_name": "Nguyen" } },
					{ "wildcard": { "code": "ORD*" } },
					{ "exists": { "field": "shipping_address" } },
					{ "range": { "amount": { "gte": 500000, "lte": 2000000 } } }
				]
			}
		},
		"sort": [
			{ "created_at": { "order": "desc" } }
		],
		"from": 0,
		"size": 25
	}`

	code, err := c.QueryToGolang(jsonInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify generated code contains user's DSL structs
	expectedSubstrings := []string{
		`Add(EsFilters{Type: "term", Key: "status", Value: "completed"})`,
		`Add(EsFilters{Type: "match", Key: "customer_name", Value: "Nguyen"})`,
		`Add(EsFilters{Type: "wildcard", Key: "code", Value: "ORD*"})`,
		`Exists("shipping_address")`,
		`Add(EsFiltersRange{Key: "amount", Min: 500000, Max: 2000000})`,
		`.Paging(0, 25)`,
		`EsSort{Key: "created_at", Order: DESC}`,
		`.Must().Bool().Query()`,
	}

	for _, exp := range expectedSubstrings {
		if !strings.Contains(code, exp) {
			t.Errorf("expected code to contain %q, but got:\n%s", exp, code)
		}
	}
}

func TestGolangToQuery_Deterministic(t *testing.T) {
	c := NewConverter(nil)

	goCode := `
	filters := EsFiltersArr{}.
		Add(EsFilters{Type: "term", Key: "order_status", Value: "paid"}).
		Add(EsFilters{Type: "match", Key: "product_title", Value: "Laptop Asus"}).
		Add(EsFilters{Type: "wildcard", Key: "sku", Value: "ASUS*"}).
		Exists("warranty_code")

	ranges := EsRangesArr{}.
		Add(EsFiltersRange{Key: "price", Min: 10000000, Max: 25000000})

	sorts := EsSortArr{}.
		Add(EsSort{Key: "price", Order: ASC})

	queryArr := append(filters.Query(), ranges.Query()...)
	query := queryArr.Must().Bool().Query().Paging(1, 20).Sort(sorts)
	`

	jsonQuery, err := c.GolangToQuery(goCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonQuery), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v\nOutput was:\n%s", err, jsonQuery)
	}

	// Verify paging
	size, hasSize := parsed["size"].(float64)
	from, hasFrom := parsed["from"].(float64)
	if !hasSize || size != 20 {
		t.Errorf("expected size 20, got %v", size)
	}
	if !hasFrom || from != 20 {
		t.Errorf("expected from 20 (page 1 * size 20), got %v", from)
	}

	// Verify sort
	sortList, ok := parsed["sort"].([]interface{})
	if !ok || len(sortList) != 1 {
		t.Fatalf("expected 1 sort item, got %v", parsed["sort"])
	}

	// Verify bool query
	queryMap, ok := parsed["query"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'query' key in output")
	}
	boolMap, ok := queryMap["bool"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'bool' in query")
	}
	mustList, ok := boolMap["must"].([]interface{})
	if !ok || len(mustList) < 5 {
		t.Errorf("expected at least 5 must items, got %d", len(mustList))
	}
}

func TestGolangToQuery_MatchAll(t *testing.T) {
	c := NewConverter(nil)

	goCode := `query := QueryMatchAll().Paging(0, 50)`
	jsonQuery, err := c.GolangToQuery(goCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonQuery), &parsed); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}

	if parsed["size"] != float64(50) {
		t.Errorf("expected size 50, got %v", parsed["size"])
	}
}
