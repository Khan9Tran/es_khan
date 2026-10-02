package ai

import (
	"strings"
	"testing"

	"eskhan/internal/schema"
)

func TestExtractQueryFields(t *testing.T) {
	queryJSON := `POST /ecommerce_orders/_search
	{
		"query": {
			"bool": {
				"must": [
					{ "term": { "order_status": "completed" } },
					{ "match": { "customer_name": "Nguyen" } },
					{ "multi_match": { "query": "laptop", "fields": ["title^3", "description"] } },
					{ "range": { "total_amount": { "gte": 1000000 } } },
					{ "exists": { "field": "shipping_address" } }
				]
			}
		},
		"sort": [
			{ "created_at": { "order": "desc" } }
		],
		"aggs": {
			"by_category": {
				"terms": {
					"field": "category_name.keyword"
				}
			}
		}
	}`

	fields, err := ExtractQueryFields(queryJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"category_name.keyword",
		"created_at",
		"customer_name",
		"description",
		"order_status",
		"shipping_address",
		"title",
		"total_amount",
	}

	for _, exp := range expected {
		found := false
		for _, f := range fields {
			if f == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected field %q to be extracted, got: %v", exp, fields)
		}
	}
}

func TestValidateQueryAgainstSchema_Valid(t *testing.T) {
	query := `{
		"query": {
			"bool": {
				"filter": [
					{ "term": { "order_status": "completed" } },
					{ "range": { "total_amount": { "gte": 500000 } } }
				]
			}
		}
	}`

	schemaFields := []schema.FieldSuggestion{
		{Name: "order_id", Type: "keyword"},
		{Name: "order_status", Type: "keyword"},
		{Name: "total_amount", Type: "double"},
		{Name: "created_at", Type: "date"},
	}

	res := ValidateQueryAgainstSchema(query, "ecommerce_orders", schemaFields)
	if !res.IsValid {
		t.Fatalf("expected query to be valid, got invalid: %v, question: %s", res.InvalidFields, res.Question)
	}
	if len(res.VerifiedFields) != 2 {
		t.Errorf("expected 2 verified fields, got %d: %v", len(res.VerifiedFields), res.VerifiedFields)
	}
}

func TestValidateQueryAgainstSchema_InvalidFields_WithSuggestions(t *testing.T) {
	// Query generated with hallucinated fields 'status' and 'price'
	query := `{
		"query": {
			"bool": {
				"filter": [
					{ "term": { "status": "completed" } },
					{ "range": { "price": { "gte": 500000 } } }
				]
			}
		}
	}`

	schemaFields := []schema.FieldSuggestion{
		{Name: "order_id", Type: "keyword"},
		{Name: "order_status", Type: "keyword", Detail: "Trạng thái đơn hàng"},
		{Name: "total_amount", Type: "double", Detail: "Tổng giá tiền"},
		{Name: "created_at", Type: "date"},
	}

	res := ValidateQueryAgainstSchema(query, "ecommerce_orders", schemaFields)
	if res.IsValid {
		t.Fatalf("expected query to be invalid due to 'status' and 'price', but it was valid")
	}

	if len(res.InvalidFields) != 2 {
		t.Errorf("expected 2 invalid fields, got %d: %v", len(res.InvalidFields), res.InvalidFields)
	}

	if !strings.Contains(res.Question, "ecommerce_orders") {
		t.Errorf("expected question to mention index name, got: %s", res.Question)
	}

	// Verify suggestions contain order_status (for status) and total_amount (for price)
	foundOrderStatus := false
	foundTotalAmount := false
	for _, opt := range res.SuggestedOptions {
		if opt.Field == "order_status" {
			foundOrderStatus = true
		}
		if opt.Field == "total_amount" {
			foundTotalAmount = true
		}
	}

	if !foundOrderStatus {
		t.Errorf("expected suggestions to contain 'order_status' for 'status', got: %+v", res.SuggestedOptions)
	}
	if !foundTotalAmount {
		t.Errorf("expected suggestions to contain 'total_amount' for 'price', got: %+v", res.SuggestedOptions)
	}
}

func TestFindClosestFields(t *testing.T) {
	schemaFields := []schema.FieldSuggestion{
		{Name: "order_id", Type: "keyword"},
		{Name: "customer_name", Type: "text"},
		{Name: "shipping_address", Type: "text"},
		{Name: "payment_status", Type: "keyword"},
		{Name: "unit_price", Type: "double"},
	}

	// Match synonym "tien" or "cost"
	costMatches := FindClosestFields("cost", schemaFields, 2)
	if len(costMatches) == 0 || costMatches[0].Name != "unit_price" {
		t.Errorf("expected unit_price for cost, got: %v", costMatches)
	}

	// Match substring / synonym "status"
	statusMatches := FindClosestFields("status", schemaFields, 2)
	if len(statusMatches) == 0 || statusMatches[0].Name != "payment_status" {
		t.Errorf("expected payment_status for status, got: %v", statusMatches)
	}
}
