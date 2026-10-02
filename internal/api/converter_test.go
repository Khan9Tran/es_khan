package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPI_ConverterEndpoints(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	// 1. Test /api/converter/to-golang
	queryPayload := map[string]string{
		"query_dsl": `POST /ecommerce_orders/_search
		{
			"query": {
				"bool": {
					"must": [
						{ "term": { "status": "completed" } }
					]
				}
			},
			"size": 20
		}`,
	}
	body, _ := json.Marshal(queryPayload)
	resp, err := http.Post(ts.URL+"/api/converter/to-golang", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/converter/to-golang failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var resGolang map[string]string
	json.NewDecoder(resp.Body).Decode(&resGolang)
	goCode := resGolang["golang_code"]
	if !strings.Contains(goCode, "EsFilters") || !strings.Contains(goCode, ".Must().Bool().Query()") {
		t.Errorf("expected generated go code to contain EsFilters and Must().Bool().Query(), got:\n%s", goCode)
	}

	// 2. Test /api/converter/to-dsl
	goPayload := map[string]string{
		"golang_code": `
		filters := EsFiltersArr{}.
			Add(EsFilters{Type: "term", Key: "order_status", Value: "paid"})
		query := filters.Query().Must().Bool().Query().Paging(0, 10)
		`,
	}
	bodyGo, _ := json.Marshal(goPayload)
	respGo, err := http.Post(ts.URL+"/api/converter/to-dsl", "application/json", bytes.NewReader(bodyGo))
	if err != nil {
		t.Fatalf("POST /api/converter/to-dsl failed: %v", err)
	}
	defer respGo.Body.Close()

	if respGo.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", respGo.StatusCode)
	}

	var resDSL map[string]string
	json.NewDecoder(respGo.Body).Decode(&resDSL)
	dslCode := resDSL["query_dsl"]
	if !strings.Contains(dslCode, "order_status") || !strings.Contains(dslCode, "paid") {
		t.Errorf("expected generated DSL to contain order_status and paid, got:\n%s", dslCode)
	}
}
