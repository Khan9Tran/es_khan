package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eskhan/internal/config"
	"eskhan/internal/storage"
)

func TestEnvironmentAPI(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. GET /api/environments
	req := httptest.NewRequest(http.MethodGet, "/api/environments", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	var listResp struct {
		Environments        []config.Environment `json:"environments"`
		ActiveEnvironmentID string               `json:"active_environment_id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(listResp.Environments) == 0 {
		t.Fatalf("expected default environment to be returned")
	}

	// 2. POST /api/environments (Create new)
	newEnv := config.Environment{
		ID:   "env-prod",
		Name: "Production",
		Variables: map[string]string{
			"base_url": "https://api.prod.com",
			"token":    "prod-secret-token",
		},
	}
	body, _ := json.Marshal(newEnv)
	req = httptest.NewRequest(http.MethodPost, "/api/environments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	// 3. POST /api/environments/env-prod/activate
	req = httptest.NewRequest(http.MethodPost, "/api/environments/env-prod/activate", nil)
	rr = httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	// 4. Verify Active Environment in list
	req = httptest.NewRequest(http.MethodGet, "/api/environments", nil)
	rr = httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	var listResp2 struct {
		ActiveEnvironmentID string `json:"active_environment_id"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&listResp2)
	if listResp2.ActiveEnvironmentID != "env-prod" {
		t.Errorf("expected active environment env-prod, got %s", listResp2.ActiveEnvironmentID)
	}

	// 5. DELETE /api/environments/env-prod
	req = httptest.NewRequest(http.MethodDelete, "/api/environments/env-prod", nil)
	rr = httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
}

func TestHTTPHistoryRecording(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// Mock target server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"hello"}`))
	}))
	defer ts.Close()

	// Send HTTP request via proxy
	sendReq := map[string]any{
		"method": "GET",
		"url":    ts.URL + "/test",
	}
	body, _ := json.Marshal(sendReq)
	req := httptest.NewRequest(http.MethodPost, "/api/http/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	// Query history with protocol=http
	histReq := httptest.NewRequest(http.MethodGet, "/api/history?protocol=http", nil)
	histRR := httptest.NewRecorder()
	srv.router.ServeHTTP(histRR, histReq)

	var histResp struct {
		History []storage.HistoryItem `json:"history"`
	}
	_ = json.NewDecoder(histRR.Body).Decode(&histResp)
	if len(histResp.History) != 1 {
		t.Fatalf("expected 1 http history item, got %d", len(histResp.History))
	}
	if histResp.History[0].Protocol != "http" || histResp.History[0].Status != 200 {
		t.Errorf("unexpected history item: %+v", histResp.History[0])
	}
}
