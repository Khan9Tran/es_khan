package es

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eskhan/internal/config"
)

func TestParseRawInput(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		defaultIndex string
		expectedM    string
		expectedP    string
		expectedBody bool
	}{
		{
			name:         "Kibana style GET with path",
			raw:          "GET _cat/indices?v",
			defaultIndex: "",
			expectedM:    "GET",
			expectedP:    "/_cat/indices?v",
			expectedBody: false,
		},
		{
			name:         "Kibana style POST with index and body",
			raw:          "POST my_index/_search\n{\n  \"query\": { \"match_all\": {} }\n}",
			defaultIndex: "",
			expectedM:    "POST",
			expectedP:    "/my_index/_search",
			expectedBody: true,
		},
		{
			name:         "Pure JSON with default index",
			raw:          "{\n  \"query\": { \"term\": { \"status\": \"active\" } }\n}",
			defaultIndex: "orders",
			expectedM:    "POST",
			expectedP:    "/orders/_search",
			expectedBody: true,
		},
		{
			name:         "Path only",
			raw:          "/_cluster/health",
			defaultIndex: "",
			expectedM:    "GET",
			expectedP:    "/_cluster/health",
			expectedBody: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, p, b := ParseRawInput(tc.raw, tc.defaultIndex)
			if m != tc.expectedM {
				t.Errorf("expected method %s, got %s", tc.expectedM, m)
			}
			if p != tc.expectedP {
				t.Errorf("expected path %s, got %s", tc.expectedP, p)
			}
			if tc.expectedBody && len(b) == 0 {
				t.Errorf("expected non-empty body")
			}
			if !tc.expectedBody && len(b) > 0 {
				t.Errorf("expected empty body, got %s", string(b))
			}
		})
	}
}

func TestExecuteQuery_Compatibility(t *testing.T) {
	// Test ES 7.x response (where total is a number)
	tsES7 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"took":      5,
			"timed_out": false,
			"_shards": map[string]interface{}{
				"total":      1,
				"successful": 1,
				"skipped":    0,
				"failed":     0,
			},
			"hits": map[string]interface{}{
				"total":     15, // Number format (ES 7.x without track_total_hits)
				"max_score": 1.0,
				"hits": []map[string]interface{}{
					{
						"_index": "products",
						"_id":    "item-1",
						"_score": 1.0,
						"_source": map[string]interface{}{
							"name":  "MacBook Pro",
							"price": 2499.0,
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tsES7.Close()

	client7 := NewClient(config.ConnectionProfile{URL: tsES7.URL})
	res7, err := client7.ExecuteQuery(context.Background(), QueryRequest{
		RawInput: "GET /products/_search",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res7.TotalHits != 15 {
		t.Errorf("expected TotalHits 15, got %d", res7.TotalHits)
	}
	if len(res7.ExtractedHits) != 1 {
		t.Fatalf("expected 1 extracted hit, got %d", len(res7.ExtractedHits))
	}
	if res7.ExtractedHits[0]["_id"] != "item-1" {
		t.Errorf("expected _id item-1, got %v", res7.ExtractedHits[0]["_id"])
	}

	// Test ES 8.x response (where total is an object: { "value": 15, "relation": "eq" })
	tsES8 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"took":      3,
			"timed_out": false,
			"_shards": map[string]interface{}{
				"total":      2,
				"successful": 2,
				"skipped":    0,
				"failed":     0,
			},
			"hits": map[string]interface{}{
				"total": map[string]interface{}{
					"value":    250,
					"relation": "gte",
				},
				"max_score": 2.5,
				"hits": []map[string]interface{}{
					{
						"_index": "logs",
						"_id":    "log-99",
						"_score": 2.5,
						"_source": map[string]interface{}{
							"level":   "ERROR",
							"message": "Out of memory",
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tsES8.Close()

	client8 := NewClient(config.ConnectionProfile{URL: tsES8.URL})
	res8, err := client8.ExecuteQuery(context.Background(), QueryRequest{
		RawInput: "POST /logs/_search\n{}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res8.TotalHits != 250 {
		t.Errorf("expected TotalHits 250, got %d", res8.TotalHits)
	}
	if res8.TotalHitsRelation != "gte" {
		t.Errorf("expected relation gte, got %s", res8.TotalHitsRelation)
	}
	if res8.TookMs != 3 {
		t.Errorf("expected took 3, got %d", res8.TookMs)
	}
}
