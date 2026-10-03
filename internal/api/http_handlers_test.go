package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIHTTPSend(t *testing.T) {
	srv, cleanupAPI := setupTestServer(t)
	defer cleanupAPI()

	// Mock target server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received":true}`))
	}))
	defer targetServer.Close()

	reqBody := map[string]any{
		"method": "POST",
		"url":    targetServer.URL + "/test",
		"headers": map[string]string{
			"Authorization": "Bearer sample_token",
		},
		"body_type": "json",
		"body":      `{"foo":"bar"}`,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/http/send", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		StatusCode int    `json:"status_code"`
		Body       string `json:"body"`
		Timings    struct {
			TotalMs int64 `json:"total_ms"`
		} `json:"timings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected target status 200, got %d", resp.StatusCode)
	}
	if resp.Body != `{"received":true}` {
		t.Errorf("unexpected body: %s", resp.Body)
	}
}

func TestAPIHTTPSend_MissingURL(t *testing.T) {
	srv, cleanupAPI := setupTestServer(t)
	defer cleanupAPI()

	reqBody := map[string]any{
		"method": "GET",
		"url":    "",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/http/send", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for empty URL, got %d", rec.Code)
	}
}
