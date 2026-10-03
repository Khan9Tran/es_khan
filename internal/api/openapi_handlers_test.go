package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIOpenAPIParse_Content(t *testing.T) {
	srv, cleanupAPI := setupTestServer(t)
	defer cleanupAPI()

	specJSON := `{
  "openapi": "3.0.0",
  "info": { "title": "My API", "version": "1.0" },
  "paths": {
    "/users": {
      "get": {
        "summary": "List users",
        "responses": { "200": { "description": "Success" } }
      }
    }
  }
}`

	reqBody := map[string]any{
		"content": specJSON,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/openapi/parse", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Info struct {
			Title string `json:"title"`
		} `json:"info"`
		Endpoints []any `json:"endpoints"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Info.Title != "My API" {
		t.Errorf("expected title 'My API', got %s", resp.Info.Title)
	}
	if len(resp.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(resp.Endpoints))
	}
}

func TestAPIOpenAPIParse_Validation(t *testing.T) {
	srv, cleanupAPI := setupTestServer(t)
	defer cleanupAPI()

	reqBody := map[string]any{
		"url":     "",
		"content": "",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/openapi/parse", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}
