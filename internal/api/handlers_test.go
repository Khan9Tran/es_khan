package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"eskhan/internal/config"
	"eskhan/internal/storage"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	tempDir, err := os.MkdirTemp("", "eskhan-api-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cfgPath := filepath.Join(tempDir, "config.json")
	cfgMgr, err := config.NewManager(cfgPath)
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}

	dataPath := filepath.Join(tempDir, "data.json")
	store, err := storage.NewStore(dataPath, 50)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	srv, err := NewServer(cfgMgr, store, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return srv, cleanup
}

func TestAPIConnections(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Get connections
	req := httptest.NewRequest(http.MethodGet, "/api/connections", nil)
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	// 2. Save a new connection
	newConn := config.ConnectionProfile{
		ID:       "test-cluster",
		Name:     "Test Cluster",
		URL:      "http://localhost:9200",
		AuthType: config.AuthNone,
	}
	body, _ := json.Marshal(newConn)
	postReq := httptest.NewRequest(http.MethodPost, "/api/connections", bytes.NewReader(body))
	postRec := httptest.NewRecorder()
	srv.Router().ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for save connection, got %d", postRec.Code)
	}

	// 3. Test snippets endpoint
	snipReq := httptest.NewRequest(http.MethodGet, "/api/snippets", nil)
	snipRec := httptest.NewRecorder()
	srv.Router().ServeHTTP(snipRec, snipReq)

	if snipRec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for snippets, got %d", snipRec.Code)
	}

	// 4. Test DSL snippets autocomplete endpoint
	dslReq := httptest.NewRequest(http.MethodGet, "/api/snippets/dsl", nil)
	dslRec := httptest.NewRecorder()
	srv.Router().ServeHTTP(dslRec, dslReq)

	if dslRec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for DSL snippets, got %d", dslRec.Code)
	}

	// 5. Test Linter check endpoint
	linterBody, _ := json.Marshal(map[string]string{
		"query": `{ "query": { "wildcard": { "name": "*sample" } } }`,
		"index": "test-index",
	})
	lintReq := httptest.NewRequest(http.MethodPost, "/api/linter/check", bytes.NewReader(linterBody))
	lintRec := httptest.NewRecorder()
	srv.Router().ServeHTTP(lintRec, lintReq)

	if lintRec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for linter, got %d", lintRec.Code)
	}

	var lintResp map[string]interface{}
	_ = json.NewDecoder(lintRec.Body).Decode(&lintResp)
	warnings, ok := lintResp["warnings"].([]interface{})
	if !ok || len(warnings) == 0 {
		t.Errorf("expected at least 1 linter warning for leading wildcard, got %v", lintResp)
	}
}

func TestAntigravityGenerate_NoIndexClarification(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	reqBody, _ := json.Marshal(map[string]string{
		"prompt": "tìm các đơn hàng đã thanh toán",
		// Index is purposely omitted
	})

	req := httptest.NewRequest(http.MethodPost, "/api/antigravity/generate", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	needsClarification, _ := resp["needs_clarification"].(bool)
	if !needsClarification {
		t.Errorf("expected needs_clarification=true when no index is provided, got: %v", resp)
	}

	question, _ := resp["question"].(string)
	if !strings.Contains(question, "Chỉ mục") && !strings.Contains(question, "Index") {
		t.Errorf("expected clarification question asking for index, got: %s", question)
	}
}

