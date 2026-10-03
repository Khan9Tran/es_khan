package openapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseOpenAPI3_JSON(t *testing.T) {
	specJSON := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Search API Service",
    "version": "1.0.0",
    "description": "Elasticsearch Gateway API"
  },
  "servers": [
    { "url": "https://api.example.com/v1" }
  ],
  "tags": [
    { "name": "Search" }
  ],
  "paths": {
    "/products/search": {
      "post": {
        "summary": "Search products in Elasticsearch",
        "tags": ["Search"],
        "parameters": [
          {
            "name": "page",
            "in": "query",
            "required": false,
            "schema": { "type": "integer", "default": 1 }
          }
        ],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/SearchQuery"
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Search results returned" }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "SearchQuery": {
        "type": "object",
        "properties": {
          "query": { "type": "string", "example": "laptop" },
          "limit": { "type": "integer", "default": 20 },
          "in_stock": { "type": "boolean" },
          "tags": {
            "type": "array",
            "items": { "type": "string" }
          }
        }
      }
    }
  }
}`

	parsed, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	if parsed.Info.Title != "Search API Service" {
		t.Errorf("expected title 'Search API Service', got %s", parsed.Info.Title)
	}
	if parsed.Info.BaseURL != "https://api.example.com/v1" {
		t.Errorf("expected base_url 'https://api.example.com/v1', got %s", parsed.Info.BaseURL)
	}

	if len(parsed.Endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(parsed.Endpoints))
	}

	ep := parsed.Endpoints[0]
	if ep.Method != "POST" || ep.Path != "/products/search" {
		t.Errorf("unexpected endpoint: %s %s", ep.Method, ep.Path)
	}

	if ep.MockBody == "" {
		t.Errorf("expected non-empty mock body generated from schema")
	}

	var mockObj map[string]any
	if err := json.Unmarshal([]byte(ep.MockBody), &mockObj); err != nil {
		t.Fatalf("failed to parse generated mock body as JSON: %v", err)
	}
	if mockObj["query"] != "laptop" {
		t.Errorf("expected mock query to be 'laptop', got %v", mockObj["query"])
	}
	if mockObj["limit"] != float64(20) {
		t.Errorf("expected mock limit to be 20, got %v", mockObj["limit"])
	}
}

func TestParseSwagger2_YAML(t *testing.T) {
	specYAML := `
swagger: "2.0"
info:
  title: "Legacy Swagger Service"
  version: "2.0.0"
host: "petstore.swagger.io"
basePath: "/v2"
schemes:
  - "https"
paths:
  /pets:
    get:
      summary: "List all pets"
      tags:
        - "Pets"
      parameters:
        - name: "limit"
          in: "query"
          type: "integer"
          required: false
      responses:
        200:
          description: "A paged array of pets"
`
	parsed, err := ParseSpec([]byte(specYAML))
	if err != nil {
		t.Fatalf("ParseSpec YAML failed: %v", err)
	}

	if parsed.Info.Title != "Legacy Swagger Service" {
		t.Errorf("expected title 'Legacy Swagger Service', got %s", parsed.Info.Title)
	}
	if parsed.Info.BaseURL != "https://petstore.swagger.io/v2" {
		t.Errorf("expected base_url 'https://petstore.swagger.io/v2', got %s", parsed.Info.BaseURL)
	}

	if len(parsed.Endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(parsed.Endpoints))
	}
	if parsed.Endpoints[0].Method != "GET" {
		t.Errorf("expected GET method, got %s", parsed.Endpoints[0].Method)
	}
}

func TestFetchAndParse_MockServer(t *testing.T) {
	specJSON := `{
  "openapi": "3.0.0",
  "info": { "title": "Remote Spec", "version": "1.0" },
  "paths": {
    "/ping": {
      "get": {
        "summary": "Ping",
        "responses": { "200": { "description": "pong" } }
      }
    }
  }
}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(specJSON))
	}))
	defer ts.Close()

	parsed, err := FetchAndParse(context.Background(), ts.URL+"/swagger.json")
	if err != nil {
		t.Fatalf("FetchAndParse failed: %v", err)
	}

	if parsed.Info.Title != "Remote Spec" {
		t.Errorf("expected Remote Spec, got %s", parsed.Info.Title)
	}
	if parsed.Info.BaseURL != ts.URL {
		t.Errorf("expected base_url to resolve to %s, got %s", ts.URL, parsed.Info.BaseURL)
	}
}
